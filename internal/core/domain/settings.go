package domain

// Settings represents a key-value configuration entry
type Settings struct {
	Key      string  `gorm:"primaryKey;type:text" json:"key"`
	Value    string  `gorm:"type:text;not null" json:"value"`
	Name     string  `gorm:"type:text" json:"name"`
	Icon     string  `gorm:"type:text" json:"icon"`
	Group    string  `gorm:"type:text;default:'General'" json:"group"`
	DataType string  `gorm:"type:text;default:'string'" json:"datatype"` // string, number, boolean, text, html
	Content  *string `gorm:"type:text" json:"content,omitempty"`         // HTML content for description/help
}

// Known setting keys
const (
	SettingSecurityEnabled       = "security_enabled"      // Enable/disable all API security
	SettingBranchID              = "branch_id"
	SettingBranchName            = "branch_name"
	SettingManagementCompany     = "management_company"
	SettingPageSize              = "page_size"
	SettingOutputFilenameFormat  = "output_filename_format"
	SettingTimeOverlap           = "time_overlap"
	SettingMaxOutputAgeDays      = "max_output_age_days"
	SettingMaxConcurrentSessions = "max_concurrent_sessions"
	SettingQueueConcurrency      = "queue_concurrency"
	SettingWALCheckpointInterval = "wal_checkpoint_interval"
	SettingWALMaxSizeMB          = "wal_max_size_mb"
)

// DefaultSettings returns the default configuration values
func DefaultSettings() []Settings {
	htmlContent := func(s string) *string { return &s }

	return []Settings{
		{Key: SettingSecurityEnabled, Value: "false", Name: "Security Enabled", Icon: "Shield", Group: "Security", DataType: "boolean", Content: htmlContent("Enable API security (authentication, HMAC). When disabled, all API endpoints are publicly accessible.")},
		{Key: SettingBranchID, Value: "001", Name: "Branch ID", Icon: "Key", Group: "General", DataType: "string", Content: htmlContent("Unique identifier for the branch.")},
		{Key: SettingBranchName, Value: "BRANCH", Name: "Branch Name", Icon: "Building", Group: "General", DataType: "string", Content: htmlContent("Display name of the branch.")},
		{Key: SettingManagementCompany, Value: "PT Company", Name: "Management Company", Icon: "Briefcase", Group: "General", DataType: "string", Content: htmlContent("Name of the management company.")},
		{Key: SettingPageSize, Value: "A4", Name: "Page Size", Icon: "FileText", Group: "PDF", DataType: "string", Content: htmlContent("Page size for the generated PDF (e.g., A4, Letter).")},
		{Key: SettingOutputFilenameFormat, Value: "{branch_id}_{date}_{gate_id}", Name: "Filename Format", Icon: "FileCode", Group: "PDF", DataType: "string", Content: htmlContent("Template for output filenames.<br>Available variables: {branch_id}, {date}, {gate_id}")},
		{Key: SettingTimeOverlap, Value: "0", Name: "Time Overlap", Icon: "Clock", Group: "Scheduling", DataType: "number", Content: htmlContent("Overlap time in minutes for scheduled tasks.")},
		{Key: SettingMaxOutputAgeDays, Value: "7", Name: "Max Output Age", Icon: "Trash2", Group: "Maintenance", DataType: "number", Content: htmlContent("Days to keep generated files before auto-deletion.")},
		{Key: SettingMaxConcurrentSessions, Value: "5", Name: "Max Sessions", Icon: "Users", Group: "Security", DataType: "number", Content: htmlContent("Maximum number of concurrent admin sessions allowed.")},
		{Key: SettingQueueConcurrency, Value: "1", Name: "Queue Concurrency", Icon: "Layers", Group: "System", DataType: "number", Content: htmlContent("Number of background workers processing the queue.")},
		{Key: SettingWALCheckpointInterval, Value: "30", Name: "WAL Checkpoint Interval", Icon: "Database", Group: "System", DataType: "number", Content: htmlContent("Interval in minutes to force a WAL checkpoint.")},
		{Key: SettingWALMaxSizeMB, Value: "20", Name: "WAL Max Size (MB)", Icon: "HardDrive", Group: "System", DataType: "number", Content: htmlContent("Maximum size of WAL file in MB before forcing checkpoint.")},
	}
}
