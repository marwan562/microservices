package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(emitCmd)
}

var emitCmd = &cobra.Command{
	Use:   "emit [event_type]",
	Short: "Emit an event to trigger flows",
	Long: `Emit an event to the Gateway's event ingestion endpoint.
This will publish the event to Redis Streams for flow processing.

Example:
  micro emit payment.completed --data '{"amount": 100}'
  micro emit user.signup -d '{"email": "test@example.com"}'`,
	Args: cobra.ExactArgs(1),
	Run:  runEmit,
}

var emitData string
var emitIdempotencyKey string

func init() {
	emitCmd.Flags().StringVarP(&emitData, "data", "d", "{}", "JSON payload for the event")
	emitCmd.Flags().StringVarP(&emitIdempotencyKey, "idempotency-key", "i", "", "Idempotency key for deduplication")
}

func runEmit(cmd *cobra.Command, args []string) {
	eventType := args[0]

	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		apiKey = os.Getenv("SAPLIY_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("Error: API key not configured. Run 'micro login' or set SAPLIY_API_KEY")
		os.Exit(1)
	}

	gatewayURL := viper.GetString("gateway_url")
	if gatewayURL == "" {
		gatewayURL = os.Getenv("SAPLIY_GATEWAY_URL")
	}
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	// Parse data
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(emitData), &payload); err != nil {
		fmt.Printf("Error: Invalid JSON data: %v\n", err)
		os.Exit(1)
	}

	// Build request body
	body := map[string]interface{}{
		"type": eventType,
		"data": payload,
	}
	bodyBytes, _ := json.Marshal(body)

	// Create request
	req, err := http.NewRequest("POST", gatewayURL+"/v1/events/emit", bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if emitIdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", emitIdempotencyKey)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)
		fmt.Printf("✓ Event emitted successfully\n")
		fmt.Printf("  Type: %s\n", eventType)
		if topic, ok := result["topic"].(string); ok {
			fmt.Printf("  Topic: %s\n", topic)
		}
		if status, ok := result["status"].(string); ok {
			fmt.Printf("  Status: %s\n", status)
		}
	} else {
		fmt.Printf("✗ Failed to emit event (HTTP %d)\n", resp.StatusCode)
		fmt.Printf("  Response: %s\n", string(respBody))
		os.Exit(1)
	}
}
