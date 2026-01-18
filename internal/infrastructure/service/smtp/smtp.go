package smtp

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
)

const message = "From: %s\nTo: %s\nSubject: Hello\n\nToken: %s\n"

// todo обернуть ошибки

type Config struct {
	Host     string `env:"SMTP_HOST"`
	Port     string `env:"SMTP_PORT"`
	Secure   string `env:"SMTP_SECURE"`
	Sender   string `env:"SMTP_SENDER"`
	Password string `env:"SMTP_PASSWORD"`
	Addr     string
}

type SenderService struct {
	cfg  *Config
	auth smtp.Auth
}

func New(cfg *Config) *SenderService {
	cfg.Addr = net.JoinHostPort(cfg.Host, cfg.Port)

	return &SenderService{
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.Sender, cfg.Password, cfg.Host),
	}
}

func (s *SenderService) Send(to string, token string) error {
	const op = "SenderService.DebugSend"

	_, err := fmt.Fprintf(os.Stdout, message, s.cfg.Sender, to, token)
	if err != nil {
		return err
	}
	return nil
}

//func (s *SenderService) Send(to string, token string) (err error) {
//	const op = "SenderService.Send"
//	client, err := s.newClient()
//	if err != nil {
//		return fmt.Errorf("%s: newClient failed: %w", op, err)
//	}
//	defer func() {
//		if err != nil {
//			_ = client.Close()
//		}
//	}()
//
//	writer, err := s.prepareWriter(client, to)
//	if err != nil {
//		return fmt.Errorf("%s: prepareWriter failed: %w", op, err)
//	}
//
//	if err = s.writeMessage(writer, to, token); err != nil {
//		return fmt.Errorf("%s: writeMessage failed: %w", op, err)
//	}
//
//	if err = client.Quit(); err != nil {
//		return fmt.Errorf("%s: client.Quit failed: %w", op, err)
//	}
//	return nil
//}

func (s *SenderService) newClient() (client *smtp.Client, err error) {
	tlsCfg := &tls.Config{ServerName: s.cfg.Host}
	conn, err := tls.Dial("tcp", s.cfg.Addr, tlsCfg)
	if err != nil {
		return nil, err
	}
	client, err = smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

func (s *SenderService) prepareWriter(client *smtp.Client, to string) (wc io.WriteCloser, err error) {
	if err = client.Auth(s.auth); err != nil {
		return nil, err
	}
	if err = client.Mail(s.cfg.Sender); err != nil {
		return nil, err
	}
	if err = client.Rcpt(to); err != nil {
		return nil, err
	}
	return client.Data()
}

func (s *SenderService) writeMessage(wc io.WriteCloser, to string, token string) error {
	msg := fmt.Sprintf(message, s.cfg.Sender, to, token)
	if _, err := wc.Write([]byte(msg)); err != nil {
		_ = wc.Close()
		return err
	}
	return wc.Close()
}
