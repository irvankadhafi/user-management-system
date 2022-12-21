package console

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
	"user-service/cacher"
	"user-service/internal/config"
	"user-service/internal/db"
	"user-service/internal/helper"
	"user-service/internal/model"
	"user-service/internal/repository"
	"user-service/rbac"
	"user-service/utils"
)

var seedCmd = &cobra.Command{
	Use:   "seeder",
	Short: "run seed-user",
	Long:  `This subcommand seeding user`,
	Run:   seeder,
}

func init() {
	RootCmd.AddCommand(seedCmd)
}

func seeder(cmd *cobra.Command, args []string) {
	// Initiate all connection like db, redis, etc
	db.InitializePostgresConn()
	generalCacher := cacher.NewCacheManager()

	redisOpts := &db.RedisConnectionPoolOptions{
		DialTimeout:     config.RedisDialTimeout(),
		ReadTimeout:     config.RedisReadTimeout(),
		WriteTimeout:    config.RedisWriteTimeout(),
		IdleCount:       config.RedisMaxIdleConn(),
		PoolSize:        config.RedisMaxActiveConn(),
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 1 * time.Minute,
	}

	redisConn, err := db.NewRedigoRedisConnectionPool(config.RedisCacheHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisConn.Close)

	redisLockConn, err := db.NewRedigoRedisConnectionPool(config.RedisLockHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisLockConn.Close)

	generalCacher.SetConnectionPool(redisConn)
	generalCacher.SetLockConnectionPool(redisLockConn)
	generalCacher.SetDefaultTTL(config.CacheTTL())

	userRepo := repository.NewUserRepository(db.PostgreSQL, generalCacher)

	userAdminCipherPwd, err := helper.HashString("123456")
	if err != nil {
		logrus.Error(err)
	}
	userAdminID := utils.GenerateID()
	userAdmin := &model.User{
		ID:          userAdminID,
		Name:        "Irvan Kadhafi",
		Email:       "irvankadhafi@mail.com",
		Password:    userAdminCipherPwd,
		Role:        rbac.RoleAdmin,
		PhoneNumber: "081927145985",
	}

	err = userRepo.Create(context.Background(), userAdmin.ID, userAdmin)
	if err != nil {
		return
	}

	userMemberCipherPwd, err := helper.HashString("123456")
	if err != nil {
		logrus.Error(err)
	}
	userMember := &model.User{
		ID:          utils.GenerateID(),
		Name:        "John Doe",
		Email:       "johndoe@mail.com",
		Password:    userMemberCipherPwd,
		Role:        rbac.RoleMember,
		PhoneNumber: "081927145985",
	}
	err = userRepo.Create(context.Background(), userAdminID, userMember)
	if err != nil {
		return
	}

	logrus.Warn("DONE!")
}
