# ADR-001: General parser architecture


## Changelog

- March 11th, 2024: Init draft;

## Status

DRAFT

## Abstract

This ADR outlines the general parser architecture designed for Juno, addressing the need for enhanced interfaces to support various chains.

## Context

Juno currently lacks the necessary architecture documents and interfaces to effectively support multiple chains. In addition, it 

## Decision
 
![Architecture](../.img/adr-001-general-architecture.png)

We decided limit Juno only responsible for being an interfaces to handle enqueuing and parsing blocks jobs. The other detailed handling tasks should be moved to Callisto, say, handling transactions, handling messages and etc.

### Enqueuer

```go
type Enqueuer interface {
    ListenAndEnqueueBlocks(queue BlockQueue) (err error)
}
```

### Block Queue

```go
type BlockQueue interface {
    EnqueueBlock(position uint64)
    
    ListenBlocks()
}
```

### Context

Context 

```go
type Context interface {
    Database() Database

    Client() Client
    
    Modules() []Module
}
```

### Client

```go
type Client interface {
    LatestBlock(position uint64) Block

    SubscribeBlocks() <-chan Block
}
```

### Database

```go
type Database interface {
    SaveBlock() error

    LatestBlock() (Block, error)

    MissingHeights(startHeight, endHeight int64) []int64
}
```

### Worker

```go
type Worker interface {
    Start(ctx Context, queue BlockQueue)
}
```

### Scheduler

```go
type PeriodicOperationsModule interface {
	// RegisterPeriodicOperations allows to register all the operations that will be run on a periodic basis.
	// The given scheduler can be used to define the periodicity of each task.
	// NOTE. This method will only be run ONCE during the module initialization.
	RegisterPeriodicOperations(scheduler Scheduler) error
}

type Scheduler interface {
    Register(module PeriodicOperationsModule)
    Start(ctx Context)
}
```

### Modules

```go
type Module interface {
    Name() string
}

type BlockModule interface {
    HandleBlock(block Block) error
}
```

### Additional operator

```go
type AdditionalOperationModule interface {
    // RegisterAdditionalOperation allows to register the operation that will be run on a periodic basis.
	// The given scheduler can be used to define the periodicity of each task.
	// NOTE. This method will only be run ONCE during the module initialization.
    RegisterAdditionalOperation(operator AdditionalOperator) error
}

type AdditionalOperator interface {
    Register(module AdditionalOperationModule)
    Start(ctx Context)
}
```



## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

- {reference link}
