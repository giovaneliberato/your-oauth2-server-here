package authorization

import (
	"goauth-extension/app/domain"
	"goauth-extension/app/domain/client"

	"github.com/spf13/viper"
)

type Service interface {
	Authorize(AuthorizationRequest) (AuthozirationContext, *domain.OAuthError)
	ApproveAuthorization(ApproveAuthorizationRequest) (AuthorizationReponse, *domain.OAuthError)
	ExchangeAuthorizationCode(ExchangeAuthorizationCodeRequest) (AuthorizationReponse, *domain.OAuthError)
}

type service struct {
	client           client.Service
	contextSigner    ContextSigner
	authorizationURL string
}

func NewService(client client.Service, signer ContextSigner) Service {
	return &service{
		client:           client,
		contextSigner:    signer,
		authorizationURL: viper.GetString("authorization.consent-url"),
	}
}

func (s *service) Authorize(request AuthorizationRequest) (AuthozirationContext, *domain.OAuthError) {
	client := s.client.GetByID(request.ClientID)

	err := Validate(client, request)
	if err != nil {
		return AuthozirationContext{}, err
	}

	ctx := AuthozirationContext{
		AuthorizationURL:           s.authorizationURL,
		ClientID:                   client.ID,
		RequestedScopes:            request.Scope,
		SignedAuthorizationContext: s.buildAuthorizationContext(request),
	}

	return ctx, nil
}

func (s *service) ApproveAuthorization(approveAuthorization ApproveAuthorizationRequest) (AuthorizationReponse, *domain.OAuthError) {
	Context, err := s.contextSigner.VerifyAndDecode(approveAuthorization.SignedAuthorizationRequest)

	if err != nil {
		return AuthorizationReponse{}, domain.InvalidApproveAuthorizationError
	}

	if !approveAuthorization.ApprovedByUser {
		resp := AuthorizationReponse{
			RedirectURI: Context.RedirectURI,
			State:       Context.State,
		}

		return resp, domain.AccessDeniedError
	}

	signedAuthorizationCode := s.buildAuthorizationCodeContext(Context, approveAuthorization)

	return AuthorizationReponse{
		SignedAuthorizationCode: signedAuthorizationCode,
		State:                   Context.State,
		RedirectURI:             Context.RedirectURI,
	}, nil
}

func (s *service) ExchangeAuthorizationCode(r ExchangeAuthorizationCodeRequest) (AuthorizationReponse, *domain.OAuthError) {
	return AuthorizationReponse{}, nil
}

func (s *service) buildAuthorizationContext(req AuthorizationRequest) string {
	Context := Context{
		ClientID:    req.ClientID,
		State:       req.State,
		Scope:       req.Scope,
		RedirectURI: req.RedirectURI,
	}

	signedContext, _ := s.contextSigner.SignAndEncode(Context)
	return signedContext
}

func (s *service) buildAuthorizationCodeContext(ctx Context, approveAuthorization ApproveAuthorizationRequest) string {
	Context := Context{
		ClientID:          ctx.ClientID,
		Scope:             ctx.Scope,
		RedirectURI:       ctx.RedirectURI,
		AuthorizationCode: approveAuthorization.AuthorizationCode,
	}

	signedContext, _ := s.contextSigner.SignAndEncode(Context)
	return signedContext
}
