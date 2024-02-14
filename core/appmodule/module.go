package appmodule

import (
	"context"

	"cosmossdk.io/core/transaction"
	"google.golang.org/grpc"
)

// AppModule is a tag interface for app module implementations to use as a basis
// for extension interfaces. It provides no functionality itself, but is the
// type that all valid app modules should provide so that they can be identified
// by other modules (usually via depinject) as app modules.
type AppModule interface {
	// IsAppModule is a dummy method to tag a struct as implementing an AppModule.
	IsAppModule()

	// IsOnePerModuleType is a dummy method to help depinject resolve modules.
	IsOnePerModuleType()
}

// HasServices is the extension interface that modules should implement to register
// implementations of services defined in .proto files.
type HasServices interface {
	AppModule

	// RegisterServices registers the module's services with the app's service
	// registrar.
	//
	// Two types of services are currently supported:
	// - read-only gRPC query services, which are the default.
	// - transaction message services, which must have the protobuf service
	//   option "cosmos.msg.v1.service" (defined in "cosmos/msg/v1/service.proto")
	//   set to true.
	//
	// The service registrar will figure out which type of service you are
	// implementing based on the presence (or absence) of protobuf options. You
	// do not need to specify this in golang code.
	RegisterServices(grpc.ServiceRegistrar) error
}

// HasMigrations is the extension interface that modules should implement to register migrations.
type HasMigrations interface {
	AppModule
	HasConsensusVersion

	// RegisterMigrations registers the module's migrations with the app's migrator.
	RegisterMigrations(MigrationRegistrar) error
}

// HasConsensusVersion is the interface for declaring a module consensus version.
type HasConsensusVersion interface {
	// ConsensusVersion is a sequence number for state-breaking change of the
	// module. It should be incremented on each consensus-breaking change
	// introduced by the module. To avoid wrong/empty versions, the initial version
	// should be set to 1.
	ConsensusVersion() uint64
}

// ResponsePreBlock represents the response from the PreBlock method.
// It can modify consensus parameters in storage and signal the caller through the return value.
// When it returns ConsensusParamsChanged=true, the caller must refresh the consensus parameter in the finalize context.
// The new context (ctx) must be passed to all the other lifecycle methods.
type ResponsePreBlock interface {
	IsConsensusParamsChanged() bool
}

// HasPreBlocker is the extension interface that modules should implement to run
// custom logic before BeginBlock.
type HasPreBlocker interface {
	AppModule
	// PreBlock is method that will be run before BeginBlock.
	PreBlock(context.Context) (ResponsePreBlock, error)
}

// HasBeginBlocker is the extension interface that modules should implement to run
// custom logic before transaction processing in a block.
type HasBeginBlocker interface {
	AppModule

	// BeginBlock is a method that will be run before transactions are processed in
	// a block.
	BeginBlock(context.Context) error
}

// HasEndBlocker is the extension interface that modules should implement to run
// custom logic after transaction processing in a block.
type HasEndBlocker interface {
	AppModule

	// EndBlock is a method that will be run after transactions are processed in
	// a block.
	EndBlock(context.Context) error
}

// HasTxValidation is the extension interface that modules should implement to run
// custom logic for validating transactions.
// It was previously known as AnteHandler/Decorator.
type HasTxValidation[T transaction.Tx] interface {
	AppModule

	// TxValidator is a method that will be run on each transaction.
	// If an error is returned:
	// 	                          ,---.
	//                           /    |
	//                          /     |
	//  You shall not pass!    /      |
	//                        /       |
	//           \       ___,'        |
	//                 <  -'          :
	//                  `-.__..--'``-,_\_
	//                     |o/ <o>` :,.)_`>
	//                     :/ `     ||/)
	//                     (_.).__,-` |\
	//                     /( `.``   `| :
	//                     \'`-.)  `  ; ;
	//                     | `       /-<
	//                     |     `  /   `.
	//     ,-_-..____     /|  `    :__..-'\
	//    /,'-.__\\  ``-./ :`      ;       \
	//    `\ `\  `\\  \ :  (   `  /  ,   `. \
	//      \` \   \\   |  | `   :  :     .\ \
	//       \ `\_  ))  :  ;     |  |      ): :
	//      (`-.-'\ ||  |\ \   ` ;  ;       | |
	//       \-_   `;;._   ( `  /  /_       | |
	//        `-.-.// ,'`-._\__/_,'         ; |
	//           \:: :     /     `     ,   /  |
	//            || |    (        ,' /   /   |
	//            ||                ,'   /    |
	TxValidator(ctx context.Context, tx T) error
}

// HasUpdateValidators is an extension interface that contains information about the AppModule and UpdateValidators.
// It can be seen as the alternative of the Cosmos SDK' HasABCIEndBlocker.
// Both are still supported.
type HasUpdateValidators interface {
	AppModule

	UpdateValidators(ctx context.Context) ([]ValidatorUpdate, error)
}

// ValidatorUpdate defines a validator update.
type ValidatorUpdate struct {
	PubKey     []byte
	PubKeyType string
	Power      int64 // updated power of the validtor
}

// **********************************************
// The following interfaces are baseapp specific and will be deprecated in the future.
// **********************************************

// HasPrepareCheckState is an extension interface that contains information about the AppModule
// and PrepareCheckState.
type HasPrepareCheckState interface {
	AppModule
	PrepareCheckState(context.Context) error
}

// HasPrecommit is an extension interface that contains information about the AppModule and Precommit.
type HasPrecommit interface {
	AppModule
	Precommit(context.Context) error
}
