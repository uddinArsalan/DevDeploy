package mapping

import (
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/handlers/dto"
	"github.com/uddinArsalan/devdeploy/internals/utils"
)

func ToEnvDomain(projectID int64, userEnvs []dto.Env) []domain.Env {
	var domainEnvs []domain.Env
	for _, env := range userEnvs {
		encryptedValue, err := utils.Encrypt(env.Value)
		if err != nil {
			continue
		}
		domainEnvs = append(domainEnvs, domain.Env{
			ID:             env.ID,
			Key:            env.Key,
			EncryptedValue: encryptedValue,
			ProjectID:      projectID,
			CreatedAt:      env.CreatedAt,
			UpdatedAt:      env.UpdatedAt,
		})
	}
	return domainEnvs
}

func ToEnvsReponse(envs []domain.Env) []dto.Env {
	var userEnvs []dto.Env
	for _, env := range envs {
		value, err := utils.Decrypt(env.EncryptedValue)
		if err != nil {
			continue
		}
		userEnvs = append(userEnvs, dto.Env{
			ID:        env.ID,
			Key:       env.Key,
			Value:     value,
			CreatedAt: env.CreatedAt,
			UpdatedAt: env.UpdatedAt,
		})
	}
	return userEnvs
}
