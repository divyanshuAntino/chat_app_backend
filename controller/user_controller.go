package controller

import (
	"net/http"
	"strings"

	"github.com/divyanshu050303/chat-app-backend/helper"
	"github.com/divyanshu050303/chat-app-backend/models"
	"github.com/divyanshu050303/chat-app-backend/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserController struct {
	Repo *repository.UserRepository
}

func (ctrl *UserController) Createuser(c *fiber.Ctx) error {
	userModel := models.UserModels{
		UserId: uuid.New().String(),
	}
	err := c.BodyParser(&userModel)
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad Request", nil)
		return err
	}
	var existingUser models.UserModels
	err = ctrl.Repo.DB.Where("user_email=?", userModel.UserEmail).Find(&userModel).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad request ", nil)
		return err
	}
	if existingUser.UserId != "" {
		helper.ApiResponse(c, http.StatusConflict, "User Allready Exist", nil)
		return nil
	}
	err = ctrl.Repo.DB.Create(&userModel).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Could not create the user", nil)
		return nil
	}
	accessToken, refreshToken, err := helper.GenerateToken(userModel)
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Could not generate the accesstoken", nil)
		return err
	}
	data := map[string]any{
		"token": map[string]string{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
		"user": map[string]any{
			"id":        userModel.UserId,
			"userName":  userModel.Name,
			"userEmail": userModel.UserEmail,
		},
	}
	helper.ApiResponse(c, http.StatusOK, "User Create successfully", data)
	return nil
}
func (ctrl *UserController) LoginUser(c *fiber.Ctx) error {
	userModel := models.UserModels{}
	err := c.BodyParser(&userModel)
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad request", nil)
		return err
	}
	var existingUser models.UserModels
	err = ctrl.Repo.DB.Where("user_email=?", userModel.UserEmail).Find(&existingUser).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad Request", nil)
		return err
	}

	if existingUser.UserId == "" {
		helper.ApiResponse(c, http.StatusNotFound, "User Not Found with this email id", nil)
		return err
	}
	if !strings.EqualFold(*existingUser.UserPassword, *userModel.UserPassword) {
		helper.ApiResponse(c, http.StatusUnauthorized, "Invalid Password", nil)
		return nil

	}
	accessToken, refreshToken, err := helper.GenerateToken(existingUser)
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Could not generate the accesstoken", nil)
		return err
	}
	data := map[string]any{
		"token": map[string]string{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
		"user": map[string]any{
			"id":        existingUser.UserId,
			"userName":  existingUser.Name,
			"userEmail": existingUser.UserEmail,
		},
	}
	helper.ApiResponse(c, http.StatusOK, "Login successfully", data)

	return nil
}

func (ctrl *UserController) UpdateUserDetails(c *fiber.Ctx) error {

	authError := helper.CheckUserIsLoggedInOrNot(c)

	if authError != "" {

		return helper.ApiResponse(c, http.StatusUnauthorized, authError, nil) // This should stop further execution
	}
	req := c.Request()
	tokenString := string(req.Header.Peek("Authorization"))
	uuid, err := helper.GetUserUUIDFromToken(tokenString)
	var existingUser models.UserModels
	err = ctrl.Repo.DB.Where("user_id=?", uuid).Find(&existingUser).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad request ", nil)
		return err
	}
	if existingUser.UserId == "" {
		helper.ApiResponse(c, http.StatusNotFound, "User Not Found", nil)
		return nil
	}
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad Request", nil)
		return nil
	}
	err = ctrl.Repo.DB.Model(&existingUser).Updates(input).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Someting went wrong", nil)
		return nil
	}
	helper.ApiResponse(c, http.StatusOK, "User updated successfully", nil)

	return nil
}
func (ctrl *UserController) GetAllUser(c *fiber.Ctx) error {
	authError := helper.CheckUserIsLoggedInOrNot(c)
	if authError != "" {
		return helper.ApiResponse(c, http.StatusUnauthorized, authError, nil)
	}
	req := c.Request()
	tokenString := string(req.Header.Peek("Authorization"))
	uuid, err := helper.GetUserUUIDFromToken(tokenString)
	var existingUser models.UserModels
	err = ctrl.Repo.DB.Where("user_id=?", uuid).Find(&existingUser).Error
	if err != nil {
		helper.ApiResponse(c, http.StatusBadRequest, "Bad request ", nil)
		return err
	}
	var users []models.UserModels
	err = ctrl.Repo.DB.Where("user_id != ?", uuid).Find(&users).Error
	if err != nil {
		return helper.ApiResponse(c, http.StatusInternalServerError, "Failed to fetch users", nil)
	}

	return helper.ApiResponse(c, http.StatusOK, "Users fetched successfully", users)
}
