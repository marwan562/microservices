package policy

import (
	"log"
	"os"
)

// NewEngine returns the configured policy engine.
// It tries to load from config/policies.json or the path in POLICY_FILE env var.
// Falls back to HardcodedPolicyEngine if JSON load fails.
func NewEngine() PolicyEngine {
	path := os.Getenv("POLICY_FILE")
	if path == "" {
		// Look in typical locations
		locations := []string{"config/policies.json", "../config/policies.json", "../../config/policies.json", "policies.json"}
		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				path = loc
				break
			}
		}
	}

	if path != "" {
		engine, err := NewJSONPolicyEngine(path)
		if err == nil {
			log.Printf("Using JSON Policy Engine with config: %s", path)
			return engine
		}
		log.Printf("Warning: Failed to load JSON policies from %s: %v. Falling back to hardcoded policies.", path, err)
	}

	log.Println("Using Hardcoded Policy Engine")
	return NewHardcodedPolicyEngine()
}
