package mediawiki

type Api interface {
	Execute(Action) error
}
