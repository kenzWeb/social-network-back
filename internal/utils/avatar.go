package utils

import (
	"math/rand"
)

var defaultAvatars = []string{
	"/uploads/avatars/default/avatar1.svg",
	"/uploads/avatars/default/avatar2.svg",
	"/uploads/avatars/default/avatar3.svg",
	"/uploads/avatars/default/avatar4.svg",
	"/uploads/avatars/default/avatar5.svg",
	"/uploads/avatars/default/avatar6.svg",
}

func GetRandomDefaultAvatar() string {
	return defaultAvatars[rand.Intn(len(defaultAvatars))]
}
