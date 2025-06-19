package main

import (
	"log"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/Hikitak/figma-comment-reporter/pkg/config"
	"github.com/Hikitak/figma-comment-reporter/pkg/email"
	"github.com/Hikitak/figma-comment-reporter/pkg/reporter"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	figmaReporter := reporter.New(
		cfg.Figma.Token,
		cfg.Figma.FileKeys,
		cfg.Report.Fields,
	)

	emailSender := email.NewSender(email.Config{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		From:         cfg.Email.From,
		To:           cfg.Email.To,
		Subject:      cfg.Email.Subject,
		Body:         cfg.Email.Body,
	})

	// Запуск по расписанию
	c := cron.New()
	c.AddFunc(cfg.Schedule, func() {
		log.Println("Generating report...")
		xlsxData, err := figmaReporter.Generate()
		if err != nil {
			log.Printf("Error generating report: %v", err)
			return
		}

		log.Println("Sending email...")
		if err := emailSender.Send(xlsxData, "figma_comments.xlsx"); err != nil {
			log.Printf("Error sending email: %v", err)
		} else {
			log.Println("Email sent successfully")
		}
	})

	c.Start()
	log.Printf("Scheduler started with cron: %s", cfg.Schedule)

	// Поддерживаем работу приложения
	select {}
}