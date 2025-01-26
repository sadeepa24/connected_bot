package constbot

func MapToSlice[T comparable, Y any](in map[T]Y) []Y {
	ot := []Y{}
	for _, val := range in {
		ot = append(ot, val)
	}
	return ot
}
func MapToSliceKey[T comparable, Y any](in map[T]Y) []T {
	ot := []T{}
	for val := range in {
		ot = append(ot, val)
	}
	return ot
}

func MapToSlicePtr[T comparable, Y any](in map[T]*Y) []Y {
	ot := []Y{}
	for _, val := range in {
		ot = append(ot, *val)
	}
	return ot
}
func MapPtrToSlicePtr[T comparable, Y any](in map[T]*Y) []*Y {
	ot := []*Y{}
	for _, val := range in {
		ot = append(ot, val)
	}
	return ot
}

func SliceToMap[T comparable, Y any](in []Y, getkey func(Y) T) map[T]Y {
	sendmap := make(map[T]Y, len(in))

	for _, val := range in {
		sendmap[getkey(val)] = val
	}

	return sendmap

}

func SliceToMapPtr[T comparable, Y any](in []Y, getkey func(Y) T) map[T]*Y {
	if in == nil {
		return map[T]*Y{}
	}

	sendmap := make(map[T]*Y, len(in))

	for i, val := range in {
		sendmap[getkey(val)] = &in[i]
	}

	return sendmap

}

func IsInSlice[T any](in []T, check func(T) bool) bool {
	for _, val := range in {
		if check(val) {
			return true

		}
	}
	return false
}

func RemoveItem[T any](in []T, docompare func(T) bool) []T {
	for i, val := range in {
		if docompare(val) {
			in = append(in[:i], in[i+1:]...)
			return in
		}
	}
	return in
}

func ExcuteMap[T comparable, Y any](in map[T]Y, excuter func(v Y, key T)) {
	for key, val := range in {
		excuter(val, key)
	}
}

func ExcuteSlice[T any](in []T, exec func(*T)) {
	for i := range in {
		exec(&in[i])
	}
}
func GetFromSlice[T any](in []T, getter func(T) bool) *T {
	for i, ttt := range in {
		if getter(ttt) {
			return &in[i]
		}
	}
	return nil
}
