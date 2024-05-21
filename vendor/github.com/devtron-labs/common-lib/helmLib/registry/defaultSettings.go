package registry

import (
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/registry"
)

type DefaultSettingsGetter interface {
	SettingsGetter
}

type DefaultSettingsGetterImpl struct {
	logger *zap.SugaredLogger
}

func NewDefaultSettingsGetter(logger *zap.SugaredLogger) *DefaultSettingsGetterImpl {
	return &DefaultSettingsGetterImpl{
		logger: logger,
	}
}

func (s *DefaultSettingsGetterImpl) GetRegistrySettings(config *Configuration) (*Settings, error) {

	registryClient, err := s.getRegistryClient(config)
	if err != nil {
		s.logger.Error("error in getting registry client", "registryUrl", config.RegistryUrl, "err", err)
		return nil, err
	}

	return &Settings{
		RegistryClient:         registryClient,
		RegistryHostURL:        config.RegistryUrl,
		RegistryConnectionType: REGISTRY_CONNECTION_TYPE_DIRECT,
		HttpClient:             nil,
		Header:                 nil,
	}, nil
}

func (s *DefaultSettingsGetterImpl) getRegistryClient(config *Configuration) (*registry.Client, error) {

	var caFilePath string
	var err error
	if len(config.RegistryCAFilePath) == 0 && config.RegistryConnectionType == SECURE_WITH_CERT {
		caFilePath, err = CreateCertificateFile(config.RegistryId, config.RegistryCertificateString)
		if err != nil {
			s.logger.Errorw("error in creating certificate file path", "registryName", config.RegistryId, "err", err)
			return nil, err
		}
	}

	config.RegistryCAFilePath = caFilePath
	httpClient, err := GetHttpClient(config)
	if err != nil {
		s.logger.Errorw("error in getting http client", "registryName", config.RegistryId, "err", err)
		return nil, err
	}

	registryClient, err := registry.NewClient(registry.ClientOptHTTPClient(httpClient))
	if err != nil {
		s.logger.Errorw("error in getting registryClient", "registryName", config.RegistryId, "err", err)
		return nil, err
	}

	if config != nil && !config.IsPublicRegistry {
		err = OCIRegistryLogin(registryClient, config)
		if err != nil {
			return nil, err
		}
	}
	return registryClient, nil
}
