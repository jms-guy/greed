package handlers

import (
	"database/sql"
	kitlog "github.com/go-kit/log"
	"github.com/jms-guy/greed/backend/api/sgrid"
	"github.com/jms-guy/greed/backend/internal/config"
	"github.com/jms-guy/greed/backend/internal/database"
	"github.com/jms-guy/greed/backend/internal/limiter"
)

//Holds state of important server structs
type AppServer struct{
	Db				*database.Queries			//SQL database queries
	Database 		*sql.DB						//SQL database 
	Config 			*config.Config				//Environment variables configured from .env file
	Logger 			kitlog.Logger				//Logging interface
	SgMail			*sgrid.SGMailService		//SendGrid mail service 
	Limiter 		*limiter.IPRateLimiter		//Rate limiter
	PClient			*plaid.APIClient
}
