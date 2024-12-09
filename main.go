package main

import (
	_ "linkz/migrations"

	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const (
	collectionName = "Linkz"
	assetPath      = "./qr-codes"
)

func generateQRCode(baseURL, srcPath, savePath string) error {
	qrc, err := qrcode.New(fmt.Sprintf("%s/%s", baseURL, srcPath))
	if err != nil {
		return fmt.Errorf("failed to generate QRCode: %w", err)
	}

	w, err := standard.New(savePath)
	if err != nil {
		return fmt.Errorf("failed to create QR writer: %w", err)
	}

	if err = qrc.Save(w); err != nil {
		return fmt.Errorf("failed to save QR code: %w", err)
	}

	return nil
}

func main() {

	app := pocketbase.New()

	// Ensure assets directory exists
	if err := os.MkdirAll(assetPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create assets directory: %v", err)
	}

	// URL redirect handler
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/{src_path}", func(e *core.RequestEvent) error {
			srcURL := e.Request.PathValue("src_path")
			record, err := app.FindFirstRecordByData(collectionName, "src", srcURL)
			if err != nil {
				return e.BadRequestError("Invalid path", err)
			}

			// Increment hits counter
			hits := record.GetInt("hits")
			record.Set("hits", hits+1)

			if err := app.Save(record); err != nil {
				log.Printf("Failed to save record: %v", err)
			}

			return e.Redirect(http.StatusPermanentRedirect, record.GetString("dst"))
		})

		return se.Next()
	})

	// Handle record creation
	app.OnRecordCreate(collectionName).BindFunc(func(e *core.RecordEvent) error {
		srcPath := e.Record.GetString("src")
		filePath := filepath.Join(assetPath, srcPath+".png")

		if err := generateQRCode(e.App.Settings().Meta.AppURL, srcPath, filePath); err != nil {
			log.Printf("Failed to generate QR code: %v", err)
			return e.Next()
		}

		file, err := filesystem.NewFileFromPath(filePath)
		if err != nil {
			log.Printf("Failed to load QR code file: %v", err)
			return e.Next()
		}

		e.Record.Set("QR_Code", file)

		return e.Next()
	})

	// Handle record updates
	app.OnRecordUpdate(collectionName).BindFunc(func(e *core.RecordEvent) error {
		srcPath := e.Record.GetString("src")
		filePath := filepath.Join(assetPath, srcPath+".png")

		if err := generateQRCode(e.App.Settings().Meta.AppURL, srcPath, filePath); err != nil {
			log.Printf("Failed to update QR code: %v", err)
			return e.Next()
		}

		file, err := filesystem.NewFileFromPath(filePath)
		if err != nil {
			log.Printf("Failed to load updated QR code file: %v", err)
			return e.Next()
		}

		e.Record.Set("QR_Code", file)

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
