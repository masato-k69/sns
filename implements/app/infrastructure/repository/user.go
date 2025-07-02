package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	irdb "app/infrastructure/adapter/datastore/rdb"
	imodel "app/infrastructure/model"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/samber/do"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	userStoreConnection irdb.UserStoreConnection
}

// List implements repository.UserRepository.
func (u *userRepository) List(c context.Context, ids []uuid.UUID) ([]dmodel.User, error) {
	users := []imodel.User{}
	if err := u.userStoreConnection.Read().
		Where("id in ?", lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })).
		Find(&users).Error; err != nil {
		return nil, errors.Wrapf(err, "fialed to list user. ids=%v", ids)
	}

	dUsers := []dmodel.User{}
	for _, user := range users {
		var imageURL *string
		if user.ImageUrl.Valid {
			imageURL = &user.ImageUrl.String
		} else {
			imageURL = nil
		}

		dUser, err := dfactory.NewUser(user.ID, user.Subject, user.Email, user.Issuer, user.Name, imageURL)

		if err != nil {
			return nil, errors.Wrapf(err, "fialed to parse user. id=%v", user.ID)
		}

		dUsers = append(dUsers, *dUser)
	}

	return dUsers, nil
}

// Get implements repository.UserRepository.
func (u *userRepository) Get(c context.Context, id uuid.UUID) (*dmodel.User, error) {
	user := imodel.User{}
	if err := u.userStoreConnection.Read().
		Where("id = ?", id.String()).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "fialed to get user. id=%v", id.String())
	}

	var imageURL *string
	if user.ImageUrl.Valid {
		imageURL = &user.ImageUrl.String
	} else {
		imageURL = nil
	}

	dUser, err := dfactory.NewUser(user.ID, user.Subject, user.Email, user.Issuer, user.Name, imageURL)

	if err != nil {
		return nil, errors.Wrapf(err, "fialed to parse user. id=%v", id.String())
	}

	return dUser, nil
}

// GetBySubject implements repository.UserRepository.
func (u *userRepository) GetBySubject(c context.Context, subject dmodel.Subject) (*dmodel.User, error) {
	user := imodel.User{}
	if err := u.userStoreConnection.Read().
		Where("subject = ?", subject).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "fialed to get user. subject=%v", subject)
	}

	var imageURL *string
	if user.ImageUrl.Valid {
		imageURL = &user.ImageUrl.String
	} else {
		imageURL = nil
	}

	dUser, err := dfactory.NewUser(user.ID, user.Subject, user.Email, user.Issuer, user.Name, imageURL)

	if err != nil {
		return nil, errors.Wrapf(err, "fialed to parse user. subject=%v", subject)
	}

	return dUser, nil
}

// Save implements repository.UserRepository.
func (u *userRepository) Save(c context.Context, user dmodel.User) error {
	return u.userStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		imageURL := func(user dmodel.User) sql.NullString {
			if user.ImageURL == nil {
				return sql.NullString{}
			}

			return sql.NullString{
				String: string(*user.ImageURL),
				Valid:  true,
			}
		}(user)

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "subject"}},
			DoUpdates: clause.AssignmentColumns([]string{"email", "image_url"}),
		}).Create(&imodel.User{
			ID:       user.ID.String(),
			Subject:  user.Subject.String(),
			Email:    user.Email.String(),
			Issuer:   user.Issuer,
			Name:     user.Name.String(),
			ImageUrl: imageURL,
		}).Error; err != nil {
			return errors.Wrapf(err, "failed to save user. id=%v", user.ID.String())
		}

		return nil
	})
}

func NewUserRepository(i *do.Injector) (drepository.UserRepository, error) {
	userStoreConnection := do.MustInvoke[irdb.UserStoreConnection](i)
	return &userRepository{
		userStoreConnection: userStoreConnection,
	}, nil
}
