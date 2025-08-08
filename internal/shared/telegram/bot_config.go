package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Custom API configs for methods not in the library

// SetMyDescriptionConfig represents a setMyDescription request
type SetMyDescriptionConfig struct {
	Description   string `json:"description,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// method returns the API method name (required by Chattable interface)
func (config SetMyDescriptionConfig) method() string {
	return "setMyDescription"
}

// params returns the parameters for the API request (required by Chattable interface)
func (config SetMyDescriptionConfig) params() (tgbotapi.Params, error) {
	params := make(tgbotapi.Params)
	if config.Description != "" {
		params["description"] = config.Description
	}
	if config.LanguageCode != "" {
		params["language_code"] = config.LanguageCode
	}
	return params, nil
}

// NewSetMyDescription creates a new setMyDescription request
func NewSetMyDescription(description, languageCode string) SetMyDescriptionConfig {
	return SetMyDescriptionConfig{
		Description:   description,
		LanguageCode: languageCode,
	}
}

// SetMyShortDescriptionConfig represents a setMyShortDescription request
type SetMyShortDescriptionConfig struct {
	ShortDescription string `json:"short_description,omitempty"`
	LanguageCode    string `json:"language_code,omitempty"`
}

// method returns the API method name (required by Chattable interface)
func (config SetMyShortDescriptionConfig) method() string {
	return "setMyShortDescription"
}

// params returns the parameters for the API request (required by Chattable interface)
func (config SetMyShortDescriptionConfig) params() (tgbotapi.Params, error) {
	params := make(tgbotapi.Params)
	if config.ShortDescription != "" {
		params["short_description"] = config.ShortDescription
	}
	if config.LanguageCode != "" {
		params["language_code"] = config.LanguageCode
	}
	return params, nil
}

// NewSetMyShortDescription creates a new setMyShortDescription request
func NewSetMyShortDescription(shortDescription, languageCode string) SetMyShortDescriptionConfig {
	return SetMyShortDescriptionConfig{
		ShortDescription: shortDescription,
		LanguageCode:    languageCode,
	}
}

// SetMyNameConfig represents a setMyName request
type SetMyNameConfig struct {
	Name         string `json:"name,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// method returns the API method name (required by Chattable interface)
func (config SetMyNameConfig) method() string {
	return "setMyName"
}

// params returns the parameters for the API request (required by Chattable interface)
func (config SetMyNameConfig) params() (tgbotapi.Params, error) {
	params := make(tgbotapi.Params)
	if config.Name != "" {
		params["name"] = config.Name
	}
	if config.LanguageCode != "" {
		params["language_code"] = config.LanguageCode
	}
	return params, nil
}

// NewSetMyName creates a new setMyName request
func NewSetMyName(name, languageCode string) SetMyNameConfig {
	return SetMyNameConfig{
		Name:         name,
		LanguageCode: languageCode,
	}
}