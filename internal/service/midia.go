package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"misaki/types"

	"github.com/wader/goutubedl"
	"go.uber.org/zap"
)

func (s *Service) DownloadYoutubeMidia(ctx context.Context, rawURL string) (*types.Midia, error) {
	data, err := url.Parse(rawURL)
	if err != nil {
		s.logger.Info("error parsing url", zap.Error(err))
		return nil, fmt.Errorf("invalid url informed")
	}

	if !s.isValidYoutubeUrl(data) {
		s.logger.Info("invalid youtube url", zap.String("url", data.String()))
		return nil, fmt.Errorf("url is not a youtube link")
	}

	midia := &types.Midia{
		Url:       data,
		OnlyAudio: false,
	}

	if err := s.getMidiaData(ctx, midia); err != nil {
		s.logger.Info("error getting midia data", zap.Error(err))
		return nil, fmt.Errorf("error getting midia data")
	}

	// TODO: Validate content size,
	// first check result.Info.Format.Filesize if exist
	// or check result.Info.Format.FilesizeApprox
	// and/or check result.Info.Duration

	if err := s.downloadMidia(ctx, midia); err != nil {
		s.logger.Info("error to download midia", zap.Error(err))
		return nil, fmt.Errorf("error downloading midia")
	}

	return midia, nil
}

func (s *Service) isValidYoutubeUrl(link *url.URL) bool {
	hostname := link.Hostname()

	if strings.Contains(hostname, "youtu.be") {
		return true
	}

	if strings.Contains(hostname, "youtube.com") &&
		strings.HasPrefix(link.RequestURI(), "/watch?v=") {
		return true
	}

	if strings.Contains(hostname, "youtube.com") &&
		strings.HasPrefix(link.RequestURI(), "/shorts/") {
		return true
	}

	return false
}

func (s *Service) getMidiaData(ctx context.Context, midia *types.Midia) error {
	result, err := goutubedl.New(ctx, midia.Url.String(), goutubedl.Options{
		Type: goutubedl.TypeSingle,
	})
	if err != nil {
		return err
	}

	midia.Data = result
	// Here we are avoiding to use result.Info.Format.Ext as extension
	// because videos with .webp extension sometimes doesn't have preview
	// in some platforms, ex Telegram
	midia.Name = fmt.Sprintf("%s.mp4", result.Info.Title)

	return nil
}

func (s *Service) downloadMidia(ctx context.Context, midia *types.Midia) error {
	content := bytes.NewBuffer([]byte{})

	result, err := midia.Data.DownloadWithOptions(context.Background(), goutubedl.DownloadOptions{
		Filter:            "best",
		DownloadAudioOnly: midia.OnlyAudio,
	})
	if err != nil {
		return err
	}
	defer result.Close()
	io.Copy(content, result)

	midia.Content = content

	return nil
}
