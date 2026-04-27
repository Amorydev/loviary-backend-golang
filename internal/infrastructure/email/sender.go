package email

// Config holds SMTP configuration
type Config struct {
    Host     string
    Port     string
    Username string
    Password string
    From     string
    Enabled  bool // if false, only log to console (dev mode)
}

// Sender interface for sending emails
type Sender interface {
    SendVerificationEmail(to, code string) error
}

// NewSender creates a new email sender
func NewSender(cfg *Config) Sender {
    return &smtpSender{
        cfg: cfg,
        log: nil, // Will be injected if needed
    }
}
