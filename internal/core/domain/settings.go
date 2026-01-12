package domain

// Settings represents a key-value configuration entry
type Settings struct {
	Key       string  `gorm:"primaryKey;type:text" json:"key"`
	Value     string  `gorm:"type:text;not null" json:"value"`
	Name      string  `gorm:"type:text" json:"name"`
	Icon      string  `gorm:"type:text" json:"icon"`
	Group     string  `gorm:"type:text;default:'General'" json:"group"`
	DataType  string  `gorm:"type:text;default:'string'" json:"datatype"` // string, number, boolean, text, html
	Content   *string `gorm:"type:text" json:"content,omitempty"`         // HTML content for description/help
	SortOrder int     `gorm:"type:int;default:0" json:"order"`
}

// Known setting keys
const (
	SettingSecurityEnabled       = "security_enabled"       // Enable/disable all API security
	SettingBranchID              = "branch_id"
	SettingBranchName            = "branch_name"
	SettingManagementCompany     = "management_company"
	SettingAnalyzerOperatorName  = "analyzer_operator_name" // Analyzer operator name for PDF header
	SettingPageSize              = "page_size"
	SettingOutputFilenameFormat  = "output_filename_format"
	SettingDataSourcePathFormat  = "datasource_path_format"
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
		// General (100)
		{SortOrder: 110, Key: SettingBranchID, Value: "001", Name: "Branch ID", Icon: "Key", Group: "General", DataType: "string", Content: htmlContent("Unique identifier for the branch.")},
		{SortOrder: 120, Key: SettingBranchName, Value: "BRANCH", Name: "Branch Name", Icon: "Building", Group: "General", DataType: "string", Content: htmlContent("Display name of the branch.")},
		{SortOrder: 130, Key: SettingManagementCompany, Value: "PT Company", Name: "Management Company", Icon: "Briefcase", Group: "General", DataType: "string", Content: htmlContent("Name of the management company.")},
		{SortOrder: 140, Key: SettingAnalyzerOperatorName, Value: "Analyzer Operator", Name: "Analyzer Operator Name", Icon: "User", Group: "General", DataType: "string", Content: htmlContent("Name of the analyzer operator displayed in PDF headers.")},

		// PDF (200)
		{SortOrder: 210, Key: SettingPageSize, Value: "A4", Name: "Page Size", Icon: "FileText", Group: "PDF", DataType: "string", Content: htmlContent("Page size for the generated PDF (e.g., A4, Letter).")},
		{SortOrder: 220, Key: SettingOutputFilenameFormat, Value: "{BranchID}_{GateID}_{DATE}", Name: "Filename Format", Icon: "FileCode", Group: "PDF", DataType: "string", Content: htmlContent("Template for output filenames.<br>Available variables: {BranchID}, {GateID}, {StationID}, {YYYY}, {YY}, {MM}, {DD}, {Date}, {Time}")},
		{SortOrder: 230, Key: SettingDataSourcePathFormat, Value: "{MM}-{YYYY}/{StationID}/{DD}{MM}{YYYY}.mdb", Name: "Data Source Path Format", Icon: "Database", Group: "PDF", DataType: "string", Content: htmlContent("Template for Access database source path.<br>Available variables: {BranchID}, {GateID}, {StationID}, {YYYY}, {YY}, {MM}, {DD}, {Date}, {Time}")},

		// Scheduling (300)
		{SortOrder: 310, Key: SettingTimeOverlap, Value: "00:00", Name: "Day Start Time", Icon: "Clock", Group: "Scheduling", DataType: "time", Content: htmlContent("Daily transaction window start time (HH:MM).<br>Example: 02:00 means transactions from 02:00 today to 01:59:59 tomorrow.")},

		// Security (400)
		{SortOrder: 410, Key: SettingSecurityEnabled, Value: "false", Name: "Security Enabled", Icon: "Shield", Group: "Security", DataType: "boolean", Content: htmlContent("Enable API security (authentication, HMAC). When disabled, all API endpoints are publicly accessible.")},
		{SortOrder: 420, Key: SettingMaxConcurrentSessions, Value: "5", Name: "Max Sessions", Icon: "Users", Group: "Security", DataType: "number", Content: htmlContent("Maximum number of concurrent admin sessions allowed.")},

		// System (500)
		{SortOrder: 510, Key: SettingQueueConcurrency, Value: "1", Name: "Queue Concurrency", Icon: "Layers", Group: "System", DataType: "number", Content: htmlContent("Number of background workers processing the queue.")},
		{SortOrder: 520, Key: SettingWALCheckpointInterval, Value: "30", Name: "WAL Checkpoint Interval", Icon: "Database", Group: "System", DataType: "number", Content: htmlContent("Interval in minutes to force a WAL checkpoint.")},
		{SortOrder: 530, Key: SettingWALMaxSizeMB, Value: "20", Name: "WAL Max Size (MB)", Icon: "HardDrive", Group: "System", DataType: "number", Content: htmlContent("Maximum size of WAL file in MB before forcing checkpoint.")},

		// Maintenance (600)
		{SortOrder: 610, Key: SettingMaxOutputAgeDays, Value: "7", Name: "Max Output Age", Icon: "Trash2", Group: "Maintenance", DataType: "number", Content: htmlContent("Days to keep generated files before auto-deletion.")},
	}
}
