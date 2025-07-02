package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewUser(id string, subject string, email string, issuer string, name string, imageURL *string) (*model.User, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedSubject, err := model.NewSubject(subject)

	if err != nil {
		return nil, err
	}

	parsedEmail, err := model.NewEmailAddress(email)

	if err != nil {
		return nil, err
	}

	parsedName, err := model.NewName(name)

	if err != nil {
		return nil, err
	}

	parsedImageURL, err := func(imageURL *string) (*model.URL, error) {
		if imageURL == nil {
			return nil, nil
		}

		return model.NewURL(*imageURL)
	}(imageURL)

	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:       parsedID,
		Subject:  *parsedSubject,
		Email:    *parsedEmail,
		Issuer:   issuer,
		Name:     *parsedName,
		ImageURL: parsedImageURL,
	}, nil
}
