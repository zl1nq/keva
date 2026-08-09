package windows

type TrayCallbacks struct {
	IconPaths []string
	Open      func()
	Lock      func()
	Quit      func()
}
