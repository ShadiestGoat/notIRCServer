package users

func BitmaskHas[T ~int | ~uint](mask T, v T) bool {
	return v&mask == v
}
