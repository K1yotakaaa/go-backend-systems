package usecase

import (
	"fmt"
	"practice-7/internal/entity"
	"practice-7/internal/usecase/repo"
	"practice-7/utils"
	"time"
)

type UserUseCase struct {
	repo *repo.UserRepo
}

func New(r *repo.UserRepo) *UserUseCase {
	return &UserUseCase{r}
}

func (u *UserUseCase) Register(user *entity.User) error {
	code := utils.GenerateCode()

	user.VerifyCode = code
	user.VerifyExpiresAt = time.Now().Add(10 * time.Minute)
	user.Verified = false

	if err := utils.SendEmail(user.Email, code); err != nil {
		fmt.Printf("Warning: Failed to send email: %v\n", err)
	}
	
	fmt.Printf("Verification code for %s: %s\n", user.Email, code)

	return u.repo.Create(user)
}

func (u *UserUseCase) Login(dto *entity.LoginUserDTO) (string, string, error) {
	user, err := u.repo.GetByUsername(dto.Username)
	if err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	if !user.Verified {
		return "", "", fmt.Errorf("please verify your email first")
	}

	if !utils.CheckPassword(user.Password, dto.Password) {
		return "", "", fmt.Errorf("invalid credentials")
	}

	access, refresh, err := utils.GenerateTokens(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	utils.RDB.Set(utils.Ctx, "auth:"+user.ID.String(), access, 15*time.Minute)
	utils.RDB.Set(utils.Ctx, "refresh:"+user.ID.String(), refresh, 7*24*time.Hour)

	return access, refresh, nil
}

func (u *UserUseCase) GetMe(id string) (*entity.User, error) {
	return u.repo.GetByID(id)
}

func (u *UserUseCase) Verify(email, code string) error {
	return u.repo.Verify(email, code)
}

func (u *UserUseCase) Promote(id string) error {
	return u.repo.Promote(id)
}