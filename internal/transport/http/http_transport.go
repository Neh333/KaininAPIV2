package httpTransport

import (
	"github.com/neh333/KaininAPIV2/internal/libs/renderer"
	"github.com/neh333/KaininAPIV2/internal/service"
	usersHttpTransport "github.com/neh333/KaininAPIV2/internal/transport/http/users"
)

type HTTPTransport struct {
	Users *usersHttpTransport.UserHandlers
	// others, ex: Orders *ordersHttpTransport.OrderHandlers
}

func NewHTTPTransport(services *service.Services, responseRenderer *renderer.ResponseRenderer) *HTTPTransport {

	return &HTTPTransport{
		Users: usersHttpTransport.NewUserHandlers(services.UserService, responseRenderer),
	}
}
