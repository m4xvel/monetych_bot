package features

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	AvatarDir        = "/app/avatars"
	MaxAvatarSize    = 5 << 20 // 5 MB
	DefaultAvatarURL = AvatarDir + "/default.jpg"
)

func (f *Features) GetUserAvatar(bot *tgbotapi.BotAPI, userID int64) string {
	photos, err := bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: userID,
		Limit:  1,
	})

	if err != nil || photos.TotalCount == 0 {
		log.Println("GetUserProfilePhotos error:", err)
		return DefaultAvatarURL
	}

	fileID := photos.Photos[0][0].FileID

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Println("GetFile error:", err)
		return DefaultAvatarURL
	}

	telegramURL := file.Link(bot.Token)

	avatarURL, err := f.downloadFile(telegramURL, userID)
	if err != nil {
		log.Println("download avatar error:", err)
		return DefaultAvatarURL
	}

	return avatarURL
}

func (f *Features) downloadFile(fileURL string, userID int64) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(parsedURL.Host, "telegram.org") {
		return "", errors.New("invalid telegram file host")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to download avatar")
	}

	limitedReader := io.LimitReader(resp.Body, MaxAvatarSize)

	if err := os.MkdirAll(AvatarDir, 0755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("user_%d.jpg", userID)
	filePath := filepath.Join(AvatarDir, fileName)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, limitedReader); err != nil {
		return "", err
	}

	return "/avatars/" + fileName, nil
}
