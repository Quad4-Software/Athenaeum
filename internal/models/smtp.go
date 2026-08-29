package models

// SMTPSettings holds outbound mail server configuration used to email
// books and Kindle "send to" deliveries.
type SMTPSettings struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	FromAddr string `json:"fromAddr"`
	UseTLS   bool   `json:"useTls"`
}

// SMTPSettingsPublic masks the password for API responses.
type SMTPSettingsPublic struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"passwordSet"`
	FromAddr    string `json:"fromAddr"`
	UseTLS      bool   `json:"useTls"`
}

// Public strips the password from the settings for client responses.
func (c SMTPSettings) Public() SMTPSettingsPublic {
	return SMTPSettingsPublic{
		Enabled:     c.Enabled,
		Host:        c.Host,
		Port:        c.Port,
		Username:    c.Username,
		PasswordSet: c.Password != "",
		FromAddr:    c.FromAddr,
		UseTLS:      c.UseTLS,
	}
}
