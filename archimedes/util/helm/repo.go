package helm

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"os"
	"sync"
)

func (h *Helm) AddRepo(name, url string) {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)

	c := repo.Entry{
		Name: name,
		URL:  url,
	}
	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		log.Print(err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		log.Print(err)
	}

	f.Update(&c)
	err = f.WriteFile(settings.RepositoryConfig, 0o644)
	if err != nil {
		return
	}

	log.Printf("%q has been added to your repositories\n", name)
}

func (h *Helm) UpdateRepos() error {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		return errors.New("no repositories found. You must add one before updating")
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, r)
	}

	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				log.Printf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				log.Printf("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()

	return nil
}

func (h *Helm) SearchRepo(name string) bool {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		return false
	}

	for _, cfg := range f.Repositories {
		if cfg.Name == name {
			return true
		}

	}

	return false
}
