package injection

import (
	anclax_app "github.com/risingwavelabs/anclax/pkg/app"
	"github.com/risingwavelabs/anclax/pkg/auth"
	taskcore "github.com/risingwavelabs/anclax/pkg/taskcore/store"
)

func InjectAuth(anclaxApp *anclax_app.Application) auth.AuthInterface {
	return anclaxApp.GetAuth()
}

func InjectTaskStore(anclaxApp *anclax_app.Application) taskcore.TaskStoreInterface {
	return anclaxApp.GetTaskStore()
}
