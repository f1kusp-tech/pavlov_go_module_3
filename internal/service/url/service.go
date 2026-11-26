package url

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"

	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrAlreadyExists = errors.New("url already exists")
)

type URLStorage interface {
	SaveURL(urlToSave, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) (int64, error)
	UpdateURL(alias, urlToSave string) error
	AliasExists(alias string) (bool, error)
}

type urlInput struct {
	URL string `json:"url" validate:"required,url"`
}

type aliasInput struct {
	Alias string `json:"alias" validate:"required,alias"`
}

type Service struct {
	storage     URLStorage
	aliasLength int
	validate    *validator.Validate
}

func New(storage URLStorage, aliasLength int) *Service {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	if err := v.RegisterValidation("alias", validateAliasTag); err != nil {
		panic("register alias validation: " + err.Error())
	}

	return &Service{
		storage:     storage,
		aliasLength: aliasLength,
		validate:    v,
	}
}

func validateAliasTag(fl validator.FieldLevel) bool {
	return isValidAlias(fl.Field().String())
}

func isValidAlias(alias string) bool {
	if alias == "" {
		return false
	}
	for _, r := range alias {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func (s *Service) validateAlias(alias string) error {
	return s.validate.Struct(aliasInput{Alias: alias})
}

func (s *Service) Save(urlToSave, alias string) (string, error) {
	if err := s.validate.Struct(urlInput{URL: urlToSave}); err != nil {
		return "", err
	}

	if alias != "" {
		if err := s.validateAlias(alias); err != nil {
			return "", err
		}

		_, err := s.storage.SaveURL(urlToSave, alias)
		if errors.Is(err, storage.ErrURLExists) {
			return "", ErrAlreadyExists
		}
		if err != nil {
			return "", err
		}

		return alias, nil
	}

	const maxAttempts = 10
	for range maxAttempts {
		alias = random.NewRandomString(s.aliasLength)

		exists, err := s.storage.AliasExists(alias)
		if err != nil {
			return "", err
		}
		if exists {
			continue
		}

		_, err = s.storage.SaveURL(urlToSave, alias)
		if errors.Is(err, storage.ErrURLExists) {
			continue
		}
		if err != nil {
			return "", err
		}

		return alias, nil
	}

	return "", fmt.Errorf("failed to generate unique alias after %d attempts", maxAttempts)
}

func (s *Service) GetURL(alias string) (string, error) {
	if err := s.validateAlias(alias); err != nil {
		return "", err
	}

	url, err := s.storage.GetURL(alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		return "", ErrNotFound
	}
	return url, err
}

func (s *Service) Delete(alias string) (int64, error) {
	if err := s.validateAlias(alias); err != nil {
		return 0, err
	}

	count, err := s.storage.DeleteURL(alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		return 0, ErrNotFound
	}
	return count, err
}

func (s *Service) Update(alias, urlToSave string) error {
	if err := s.validateAlias(alias); err != nil {
		return err
	}
	if err := s.validate.Struct(urlInput{URL: urlToSave}); err != nil {
		return err
	}

	err := s.storage.UpdateURL(alias, urlToSave)
	if errors.Is(err, storage.ErrURLNotFound) {
		return ErrNotFound
	}
	return err
}

func AliasLengthFromEnv(aliasLengthStr string) (int, error) {
	if aliasLengthStr == "" {
		return 5, nil
	}
	n, err := strconv.Atoi(aliasLengthStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ALIAS_LENGTH: %w", err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("ALIAS_LENGTH must be positive, got %d", n)
	}
	return n, nil
}