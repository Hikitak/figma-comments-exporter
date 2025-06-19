package config

type Config struct {
	Figma    FigmaConfig  `yaml:"figma"`
	Schedule string       `yaml:"schedule"`
	Email    EmailConfig  `yaml:"email"`
	Report   ReportConfig `yaml:"report"`
}

type FigmaConfig struct {
	Token    string   `yaml:"token"`
	FileKeys []string `yaml:"file_keys"`
}

type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host"`
	SMTPPort     int      `yaml:"smtp_port"`
	SMTPUsername string   `yaml:"smtp_username"`
	SMTPPassword string   `yaml:"smtp_password"`
	From         string   `yaml:"from"`
	To           []string `yaml:"to"`
	Subject      string   `yaml:"subject"`
	Body         string   `yaml:"body"`
}

type ReportField struct {
	Name    string `yaml:"name"`
	Display string `yaml:"display"`
	Format  string `yaml:"format,omitempty"`
}

type ReportConfig struct {
	Fields []ReportField `yaml:"fields"`
}