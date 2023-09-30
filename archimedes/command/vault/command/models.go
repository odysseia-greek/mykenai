package command

import (
	"encoding/json"
	"time"
)

func UnmarshalClusterKeys(data []byte) (ClusterKeys, error) {
	var r ClusterKeys
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ClusterKeys) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type ClusterKeys struct {
	UnsealKeysB64         []string      `json:"unseal_keys_b64"`
	UnsealKeysHex         []string      `json:"unseal_keys_hex"`
	UnsealShares          int64         `json:"unseal_shares"`
	UnsealThreshold       int64         `json:"unseal_threshold"`
	RecoveryKeysB64       []interface{} `json:"recovery_keys_b64"`
	RecoveryKeysHex       []interface{} `json:"recovery_keys_hex"`
	RecoveryKeysShares    int64         `json:"recovery_keys_shares"`
	RecoveryKeysThreshold int64         `json:"recovery_keys_threshold"`
	RootToken             string        `json:"root_token"`
}

func UnmarshalStatus(data []byte) (Status, error) {
	var r Status
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Status) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Status struct {
	Type         string    `json:"type"`
	Initialized  bool      `json:"initialized"`
	Sealed       bool      `json:"sealed"`
	T            int       `json:"t"`
	N            int       `json:"n"`
	Progress     int       `json:"progress"`
	Nonce        string    `json:"nonce"`
	Version      string    `json:"version"`
	BuildDate    time.Time `json:"build_date"`
	Migration    bool      `json:"migration"`
	RecoverySeal bool      `json:"recovery_seal"`
	StorageType  string    `json:"storage_type"`
	HaEnabled    bool      `json:"ha_enabled"`
	ActiveTime   time.Time `json:"active_time"`
}

type vaultConfig struct {
	HaEnabled      bool
	PrimaryNode    string
	SecondaryNodes []string
}
