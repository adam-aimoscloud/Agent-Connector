package internal

import (
	"time"

	"gorm.io/gorm"
)

// AgentType agent platform type enum
type AgentType string

const (
	AgentTypeDify             AgentType = "dify"
	AgentTypeOpenAI           AgentType = "openai"
	AgentTypeOpenAICompatible AgentType = "openai_compatible"
)

// SystemConfig system configuration table
type SystemConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Agent agent configuration table
type Agent struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string         `json:"name" gorm:"type:varchar(255);not null;comment:'agent name'"`
	Type             AgentType      `json:"type" gorm:"type:varchar(50);not null;comment:'agent type: dify, openai, openai_compatible'"`
	URL              string         `json:"url" gorm:"type:varchar(500);not null;comment:'agent url'"`
	SourceAPIKey     string         `json:"source_api_key" gorm:"type:varchar(500);not null;comment:'source api key'"`
	ConnectorAPIKey  string         `json:"connector_api_key" gorm:"type:varchar(500);not null;unique;comment:'connector api key, used for data flow api authentication'"`
	AgentID          string         `json:"agent_id" gorm:"type:varchar(100);not null;unique;comment:'agent id'"`
	QPS              int            `json:"qps" gorm:"type:int;not null;default:10;comment:'agent qps limit'"`
	Enabled          bool           `json:"enabled" gorm:"type:boolean;not null;default:true;comment:'whether to enable'"`
	Description      string         `json:"description" gorm:"type:text;comment:'description'"`
	SupportStreaming bool           `json:"support_streaming" gorm:"type:boolean;not null;default:true;comment:'whether to support streaming response'"`
	ResponseFormat   string         `json:"response_format" gorm:"type:varchar(50);not null;default:'openai';comment:'response format: openai or dify'"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specify table name
func (Agent) TableName() string {
	return "agents"
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
