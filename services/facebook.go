package services

type FacebookService struct {
	AccessToken string
}

func CreateFacebookService(accessToken string) *FacebookService {
	return &FacebookService{
		AccessToken: accessToken,
	}
}

func (fs *FacebookService) ScanGroup(groupId string) ([]string, error) {
	return []string{"post1", "post2"}, nil
}