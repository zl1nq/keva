package windows

type TrayCallbacks struct {
	Open func()
	Lock func()
	Quit func()
}
