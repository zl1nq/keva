package security

func ZeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
