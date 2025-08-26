package eventbus

// Minimal global event emitter wrapper. Set once at startup.

var emitter func(name string, data any) = func(string, any) {}

func SetEmitter(f func(name string, data any)) {
	if f == nil {
		return
	}
	emitter = f
}

func Emit(name string, data any) {
	emitter(name, data)
}
