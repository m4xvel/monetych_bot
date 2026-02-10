package features

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/apperr"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

const (
	AvatarDir        = "/avatars"
	MaxAvatarSize    = 5 << 20 // 5 MB
	DefaultAvatarURL = AvatarDir + "/default.png"
)

func (f *Features) GetUserAvatar(bot *tgbotapi.BotAPI, chatID int64) string {
	photos, err := bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: chatID,
		Limit:  1,
	})

	if err != nil || photos.TotalCount == 0 {
		if err != nil {
			err = apperr.WrapTelegram("telegram.get_user_profile_photos", err)
		}
		logger.Log.Warnw("user photo not found",
			"chat_id", chatID,
			"err", err,
		)
		return DefaultAvatarURL
	}

	fileID := photos.Photos[0][0].FileID

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		err = apperr.WrapTelegram("telegram.get_file", err)
		logger.Log.Errorw("failed to get photo file",
			"chat_id", chatID,
			"err", err,
		)
		return DefaultAvatarURL
	}

	telegramURL := file.Link(bot.Token)

	avatarURL, err := f.downloadFile(telegramURL, chatID)
	if err != nil {
		logger.Log.Errorw("failed to download photo file",
			"chat_id", chatID,
			"err", err,
		)
		return DefaultAvatarURL
	}

	logger.Log.Infow("photo successfully download",
		"chat_id", chatID,
	)

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
