package observer

type Observer interface{
	Notify(data interface{}) 
}