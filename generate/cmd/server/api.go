package api

import "errors"
import "net/http"
import "github.com/marshome/apis/restful"

var _ = errors.New("")
var _ = http.DefaultClient

// AllocateIdsRequest: The request for Datastore.AllocateIds.
type AllocateIdsRequest struct {
	ProjectId string `json:"projectId"`
	// Keys: A list of keys with incomplete key paths for which to allocate
	// IDs.
	// No key may be reserved/read-only.
	Keys []*Key `json:"keys,omitempty"`
}

// AllocateIdsResponse: The response for Datastore.AllocateIds.
type AllocateIdsResponse struct {
	// Keys: The keys specified in the request (in the same order), each
	// with
	// its key path completed with a newly allocated ID.
	Keys []*Key `json:"keys,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// ArrayValue: An array value.
type ArrayValue struct {
	// Values: Values in the array.
	// The order of this array may not be preserved if it contains a mix
	// of
	// indexed and unindexed values.
	Values []*Value `json:"values,omitempty"`
}

// BeginTransactionRequest: The request for Datastore.BeginTransaction.
type BeginTransactionRequest struct {
	ProjectId string `json:"projectId"`
}

// BeginTransactionResponse: The response for
// Datastore.BeginTransaction.
type BeginTransactionResponse struct {
	// Transaction: The transaction identifier (always present).
	Transaction string `json:"transaction,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// CommitRequest: The request for Datastore.Commit.
type CommitRequest struct {
	ProjectId string `json:"projectId"`
	// Mode: The type of commit to perform. Defaults to `TRANSACTIONAL`.
	//
	// Possible values:
	//   "MODE_UNSPECIFIED" - Unspecified. This value must not be used.
	//   "TRANSACTIONAL" - Transactional: The mutations are either all
	// applied, or none are applied.
	// Learn about transactions
	// [here](https://cloud.google.com/datastore/docs/concepts/transactions).
	//   "NON_TRANSACTIONAL" - Non-transactional: The mutations may not
	// apply as all or none.
	Mode string `json:"mode,omitempty"`

	// Mutations: The mutations to perform.
	//
	// When mode is `TRANSACTIONAL`, mutations affecting a single entity
	// are
	// applied in order. The following sequences of mutations affecting a
	// single
	// entity are not permitted in a single `Commit` request:
	//
	// - `insert` followed by `insert`
	// - `update` followed by `insert`
	// - `upsert` followed by `insert`
	// - `delete` followed by `update`
	//
	// When mode is `NON_TRANSACTIONAL`, no two mutations may affect a
	// single
	// entity.
	Mutations []*Mutation `json:"mutations,omitempty"`

	// Transaction: The identifier of the transaction associated with the
	// commit. A
	// transaction identifier is returned by a call
	// to
	// Datastore.BeginTransaction.
	Transaction string `json:"transaction,omitempty"`
}

// CommitResponse: The response for Datastore.Commit.
type CommitResponse struct {
	// IndexUpdates: The number of index entries updated during the commit,
	// or zero if none were
	// updated.
	IndexUpdates int64 `json:"indexUpdates,omitempty"`

	// MutationResults: The result of performing the mutations.
	// The i-th mutation result corresponds to the i-th mutation in the
	// request.
	MutationResults []*MutationResult `json:"mutationResults,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// CompositeFilter: A filter that merges multiple other filters using
// the given operator.
type CompositeFilter struct {
	// Filters: The list of filters to combine.
	// Must contain at least one filter.
	Filters []*Filter `json:"filters,omitempty"`

	// Op: The operator for combining multiple filters.
	//
	// Possible values:
	//   "OPERATOR_UNSPECIFIED" - Unspecified. This value must not be used.
	//   "AND" - The results are required to satisfy each of the combined
	// filters.
	Op string `json:"op,omitempty"`
}

// Entity: A Datastore data object.
//
// An entity is limited to 1 megabyte when stored. That
// _roughly_
// corresponds to a limit of 1 megabyte for the serialized form of
// this
// message.
type Entity struct {
	// Key: The entity's key.
	//
	// An entity must have a key, unless otherwise documented (for
	// example,
	// an entity in `Value.entity_value` may have no key).
	// An entity's kind is its key path's last element's kind,
	// or null if it has no key.
	Key *Key `json:"key,omitempty"`

	// Properties: The entity's properties.
	// The map's keys are property names.
	// A property name matching regex `__.*__` is reserved.
	// A reserved property name is forbidden in certain documented
	// contexts.
	// The name must not contain more than 500 characters.
	// The name cannot be "".
	Properties map[string]Value `json:"properties,omitempty"`
}

// EntityResult: The result of fetching an entity from Datastore.
type EntityResult struct {
	// Cursor: A cursor that points to the position after the result
	// entity.
	// Set only when the `EntityResult` is part of a `QueryResultBatch`
	// message.
	Cursor string `json:"cursor,omitempty"`

	// Entity: The resulting entity.
	Entity *Entity `json:"entity,omitempty"`

	// Version: The version of the entity, a strictly positive number that
	// monotonically
	// increases with changes to the entity.
	//
	// This field is set for `FULL` entity
	// results.
	//
	// For missing entities in `LookupResponse`, this
	// is the version of the snapshot that was used to look up the entity,
	// and it
	// is always set except for eventually consistent reads.
	Version int64 `json:"version,omitempty,string"`
}

// Filter: A holder for any type of filter.
type Filter struct {
	// CompositeFilter: A composite filter.
	CompositeFilter *CompositeFilter `json:"compositeFilter,omitempty"`

	// PropertyFilter: A filter on a property.
	PropertyFilter *PropertyFilter `json:"propertyFilter,omitempty"`
}

// GqlQuery: A [GQL
// query](https://cloud.google.com/datastore/docs/apis/gql/gql_reference)
// .
type GqlQuery struct {
	// AllowLiterals: When false, the query string must not contain any
	// literals and instead must
	// bind all values. For example,
	// `SELECT * FROM Kind WHERE a = 'string literal'` is not allowed,
	// while
	// `SELECT * FROM Kind WHERE a = @value` is.
	AllowLiterals bool `json:"allowLiterals,omitempty"`

	// NamedBindings: For each non-reserved named binding site in the query
	// string, there must be
	// a named parameter with that name, but not necessarily the
	// inverse.
	//
	// Key must match regex `A-Za-z_$*`, must not match regex
	// `__.*__`, and must not be "".
	NamedBindings map[string]GqlQueryParameter `json:"namedBindings,omitempty"`

	// PositionalBindings: Numbered binding site @1 references the first
	// numbered parameter,
	// effectively using 1-based indexing, rather than the usual 0.
	//
	// For each binding site numbered i in `query_string`, there must be an
	// i-th
	// numbered parameter. The inverse must also be true.
	PositionalBindings []*GqlQueryParameter `json:"positionalBindings,omitempty"`

	// QueryString: A string of the format
	// described
	// [here](https://cloud.google.com/datastore/docs/apis/gql/gql_
	// reference).
	QueryString string `json:"queryString,omitempty"`
}

// GqlQueryParameter: A binding parameter for a GQL query.
type GqlQueryParameter struct {
	// Cursor: A query cursor. Query cursors are returned in query
	// result batches.
	Cursor string `json:"cursor,omitempty"`

	// Value: A value parameter.
	Value *Value `json:"value,omitempty"`
}

// Key: A unique identifier for an entity.
// If a key's partition ID or any of its path kinds or names
// are
// reserved/read-only, the key is reserved/read-only.
// A reserved/read-only key is forbidden in certain documented contexts.
type Key struct {
	// PartitionId: Entities are partitioned into subsets, currently
	// identified by a project
	// ID and namespace ID.
	// Queries are scoped to a single partition.
	PartitionId *PartitionId `json:"partitionId,omitempty"`

	// Path: The entity path.
	// An entity path consists of one or more elements composed of a kind
	// and a
	// string or numerical identifier, which identify entities. The
	// first
	// element identifies a _root entity_, the second element identifies
	// a _child_ of the root entity, the third element identifies a child of
	// the
	// second entity, and so forth. The entities identified by all prefixes
	// of
	// the path are called the element's _ancestors_.
	//
	// An entity path is always fully complete: *all* of the entity's
	// ancestors
	// are required to be in the path along with the entity identifier
	// itself.
	// The only exception is that in some documented cases, the identifier
	// in the
	// last path element (for the entity) itself may be omitted. For
	// example,
	// the last path element of the key of `Mutation.insert` may have
	// no
	// identifier.
	//
	// A path can never be empty, and a path can have at most 100 elements.
	Path []*PathElement `json:"path,omitempty"`
}

// KindExpression: A representation of a kind.
type KindExpression struct {
	// Name: The name of the kind.
	Name string `json:"name,omitempty"`
}

// LatLng: An object representing a latitude/longitude pair. This is
// expressed as a pair
// of doubles representing degrees latitude and degrees longitude.
// Unless
// specified otherwise, this must conform to the
// <a
// href="http://www.unoosa.org/pdf/icg/2012/template/WGS_84.pdf">WGS84
// st
// andard</a>. Values must be within normalized ranges.
//
// Example of normalization code in Python:
//
//     def NormalizeLongitude(longitude):
//       """Wraps decimal degrees longitude to [-180.0, 180.0]."""
//       q, r = divmod(longitude, 360.0)
//       if r > 180.0 or (r == 180.0 and q <= -1.0):
//         return r - 360.0
//       return r
//
//     def NormalizeLatLng(latitude, longitude):
//       """Wraps decimal degrees latitude and longitude to
//       [-90.0, 90.0] and [-180.0, 180.0], respectively."""
//       r = latitude % 360.0
//       if r <= 90.0:
//         return r, NormalizeLongitude(longitude)
//       elif r >= 270.0:
//         return r - 360, NormalizeLongitude(longitude)
//       else:
//         return 180 - r, NormalizeLongitude(longitude + 180.0)
//
//     assert 180.0 == NormalizeLongitude(180.0)
//     assert -180.0 == NormalizeLongitude(-180.0)
//     assert -179.0 == NormalizeLongitude(181.0)
//     assert (0.0, 0.0) == NormalizeLatLng(360.0, 0.0)
//     assert (0.0, 0.0) == NormalizeLatLng(-360.0, 0.0)
//     assert (85.0, 180.0) == NormalizeLatLng(95.0, 0.0)
//     assert (-85.0, -170.0) == NormalizeLatLng(-95.0, 10.0)
//     assert (90.0, 10.0) == NormalizeLatLng(90.0, 10.0)
//     assert (-90.0, -10.0) == NormalizeLatLng(-90.0, -10.0)
//     assert (0.0, -170.0) == NormalizeLatLng(-180.0, 10.0)
//     assert (0.0, -170.0) == NormalizeLatLng(180.0, 10.0)
//     assert (-90.0, 10.0) == NormalizeLatLng(270.0, 10.0)
//     assert (90.0, 10.0) == NormalizeLatLng(-270.0, 10.0)
type LatLng struct {
	// Latitude: The latitude in degrees. It must be in the range [-90.0,
	// +90.0].
	Latitude float64 `json:"latitude,omitempty"`

	// Longitude: The longitude in degrees. It must be in the range [-180.0,
	// +180.0].
	Longitude float64 `json:"longitude,omitempty"`
}

// LookupRequest: The request for Datastore.Lookup.
type LookupRequest struct {
	ProjectId string `json:"projectId"`
	// Keys: Keys of entities to look up.
	Keys []*Key `json:"keys,omitempty"`

	// ReadOptions: The options for this lookup request.
	ReadOptions *ReadOptions `json:"readOptions,omitempty"`
}

// LookupResponse: The response for Datastore.Lookup.
type LookupResponse struct {
	// Deferred: A list of keys that were not looked up due to resource
	// constraints. The
	// order of results in this field is undefined and has no relation to
	// the
	// order of the keys in the input.
	Deferred []*Key `json:"deferred,omitempty"`

	// Found: Entities found as `ResultType.FULL` entities. The order of
	// results in this
	// field is undefined and has no relation to the order of the keys in
	// the
	// input.
	Found []*EntityResult `json:"found,omitempty"`

	// Missing: Entities not found as `ResultType.KEY_ONLY` entities. The
	// order of results
	// in this field is undefined and has no relation to the order of the
	// keys
	// in the input.
	Missing []*EntityResult `json:"missing,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// Mutation: A mutation to apply to an entity.
type Mutation struct {
	// BaseVersion: The version of the entity that this mutation is being
	// applied to. If this
	// does not match the current version on the server, the mutation
	// conflicts.
	BaseVersion int64 `json:"baseVersion,omitempty,string"`

	// Delete: The key of the entity to delete. The entity may or may not
	// already exist.
	// Must have a complete key path and must not be reserved/read-only.
	Delete *Key `json:"delete,omitempty"`

	// Insert: The entity to insert. The entity must not already exist.
	// The entity key's final path element may be incomplete.
	Insert *Entity `json:"insert,omitempty"`

	// Update: The entity to update. The entity must already exist.
	// Must have a complete key path.
	Update *Entity `json:"update,omitempty"`

	// Upsert: The entity to upsert. The entity may or may not already
	// exist.
	// The entity key's final path element may be incomplete.
	Upsert *Entity `json:"upsert,omitempty"`
}

// MutationResult: The result of applying a mutation.
type MutationResult struct {
	// ConflictDetected: Whether a conflict was detected for this mutation.
	// Always false when a
	// conflict detection strategy field is not set in the mutation.
	ConflictDetected bool `json:"conflictDetected,omitempty"`

	// Key: The automatically allocated key.
	// Set only when the mutation allocated a key.
	Key *Key `json:"key,omitempty"`

	// Version: The version of the entity on the server after processing the
	// mutation. If
	// the mutation doesn't change anything on the server, then the version
	// will
	// be the version of the current entity or, if no entity is present, a
	// version
	// that is strictly greater than the version of any previous entity and
	// less
	// than the version of any possible future entity.
	Version int64 `json:"version,omitempty,string"`
}

// PartitionId: A partition ID identifies a grouping of entities. The
// grouping is always
// by project and namespace, however the namespace ID may be empty.
//
// A partition ID contains several dimensions:
// project ID and namespace ID.
//
// Partition dimensions:
//
// - May be "".
// - Must be valid UTF-8 bytes.
// - Must have values that match regex `[A-Za-z\d\.\-_]{1,100}`
// If the value of any dimension matches regex `__.*__`, the partition
// is
// reserved/read-only.
// A reserved/read-only partition ID is forbidden in certain
// documented
// contexts.
//
// Foreign partition IDs (in which the project ID does
// not match the context project ID ) are discouraged.
// Reads and writes of foreign partition IDs may fail if the project is
// not in an active state.
type PartitionId struct {
	// NamespaceId: If not empty, the ID of the namespace to which the
	// entities belong.
	NamespaceId string `json:"namespaceId,omitempty"`

	// ProjectId: The ID of the project to which the entities belong.
	ProjectId string `json:"projectId,omitempty"`
}

// PathElement: A (kind, ID/name) pair used to construct a key path.
//
// If either name or ID is set, the element is complete.
// If neither is set, the element is incomplete.
type PathElement struct {
	// Id: The auto-allocated ID of the entity.
	// Never equal to zero. Values less than zero are discouraged and may
	// not
	// be supported in the future.
	Id int64 `json:"id,omitempty,string"`

	// Kind: The kind of the entity.
	// A kind matching regex `__.*__` is reserved/read-only.
	// A kind must not contain more than 1500 bytes when UTF-8
	// encoded.
	// Cannot be "".
	Kind string `json:"kind,omitempty"`

	// Name: The name of the entity.
	// A name matching regex `__.*__` is reserved/read-only.
	// A name must not be more than 1500 bytes when UTF-8 encoded.
	// Cannot be "".
	Name string `json:"name,omitempty"`
}

// Projection: A representation of a property in a projection.
type Projection struct {
	// Property: The property to project.
	Property *PropertyReference `json:"property,omitempty"`
}

// PropertyFilter: A filter on a specific property.
type PropertyFilter struct {
	// Op: The operator to filter by.
	//
	// Possible values:
	//   "OPERATOR_UNSPECIFIED" - Unspecified. This value must not be used.
	//   "LESS_THAN" - Less than.
	//   "LESS_THAN_OR_EQUAL" - Less than or equal.
	//   "GREATER_THAN" - Greater than.
	//   "GREATER_THAN_OR_EQUAL" - Greater than or equal.
	//   "EQUAL" - Equal.
	//   "HAS_ANCESTOR" - Has ancestor.
	Op string `json:"op,omitempty"`

	// Property: The property to filter by.
	Property *PropertyReference `json:"property,omitempty"`

	// Value: The value to compare the property to.
	Value *Value `json:"value,omitempty"`
}

// PropertyOrder: The desired order for a specific property.
type PropertyOrder struct {
	// Direction: The direction to order by. Defaults to `ASCENDING`.
	//
	// Possible values:
	//   "DIRECTION_UNSPECIFIED" - Unspecified. This value must not be used.
	//   "ASCENDING" - Ascending.
	//   "DESCENDING" - Descending.
	Direction string `json:"direction,omitempty"`

	// Property: The property to order by.
	Property *PropertyReference `json:"property,omitempty"`
}

// PropertyReference: A reference to a property relative to the kind
// expressions.
type PropertyReference struct {
	// Name: The name of the property.
	// If name includes "."s, it may be interpreted as a property name path.
	Name string `json:"name,omitempty"`
}

// Query: A query for entities.
type Query struct {
	// DistinctOn: The properties to make distinct. The query results will
	// contain the first
	// result for each distinct combination of values for the given
	// properties
	// (if empty, all results are returned).
	DistinctOn []*PropertyReference `json:"distinctOn,omitempty"`

	// EndCursor: An ending point for the query results. Query cursors
	// are
	// returned in query result batches and
	// [can only be used to limit the same
	// query](https://cloud.google.com/datastore/docs/concepts/queries#cursor
	// s_limits_and_offsets).
	EndCursor string `json:"endCursor,omitempty"`

	// Filter: The filter to apply.
	Filter *Filter `json:"filter,omitempty"`

	// Kind: The kinds to query (if empty, returns entities of all
	// kinds).
	// Currently at most 1 kind may be specified.
	Kind []*KindExpression `json:"kind,omitempty"`

	// Limit: The maximum number of results to return. Applies after all
	// other
	// constraints. Optional.
	// Unspecified is interpreted as no limit.
	// Must be >= 0 if specified.
	Limit int64 `json:"limit,omitempty"`

	// Offset: The number of results to skip. Applies before limit, but
	// after all other
	// constraints. Optional. Must be >= 0 if specified.
	Offset int64 `json:"offset,omitempty"`

	// Order: The order to apply to the query results (if empty, order is
	// unspecified).
	Order []*PropertyOrder `json:"order,omitempty"`

	// Projection: The projection to return. Defaults to returning all
	// properties.
	Projection []*Projection `json:"projection,omitempty"`

	// StartCursor: A starting point for the query results. Query cursors
	// are
	// returned in query result batches and
	// [can only be used to continue the same
	// query](https://cloud.google.com/datastore/docs/concepts/queries#cursor
	// s_limits_and_offsets).
	StartCursor string `json:"startCursor,omitempty"`
}

// QueryResultBatch: A batch of results produced by a query.
type QueryResultBatch struct {
	// EndCursor: A cursor that points to the position after the last result
	// in the batch.
	EndCursor string `json:"endCursor,omitempty"`

	// EntityResultType: The result type for every entity in
	// `entity_results`.
	//
	// Possible values:
	//   "RESULT_TYPE_UNSPECIFIED" - Unspecified. This value is never used.
	//   "FULL" - The key and properties.
	//   "PROJECTION" - A projected subset of properties. The entity may
	// have no key.
	//   "KEY_ONLY" - Only the key.
	EntityResultType string `json:"entityResultType,omitempty"`

	// EntityResults: The results for this batch.
	EntityResults []*EntityResult `json:"entityResults,omitempty"`

	// MoreResults: The state of the query after the current batch.
	//
	// Possible values:
	//   "MORE_RESULTS_TYPE_UNSPECIFIED" - Unspecified. This value is never
	// used.
	//   "NOT_FINISHED" - There may be additional batches to fetch from this
	// query.
	//   "MORE_RESULTS_AFTER_LIMIT" - The query is finished, but there may
	// be more results after the limit.
	//   "MORE_RESULTS_AFTER_CURSOR" - The query is finished, but there may
	// be more results after the end
	// cursor.
	//   "NO_MORE_RESULTS" - The query has been exhausted.
	MoreResults string `json:"moreResults,omitempty"`

	// SkippedCursor: A cursor that points to the position after the last
	// skipped result.
	// Will be set when `skipped_results` != 0.
	SkippedCursor string `json:"skippedCursor,omitempty"`

	// SkippedResults: The number of results skipped, typically because of
	// an offset.
	SkippedResults int64 `json:"skippedResults,omitempty"`

	// SnapshotVersion: The version number of the snapshot this batch was
	// returned from.
	// This applies to the range of results from the query's `start_cursor`
	// (or
	// the beginning of the query if no cursor was given) to this
	// batch's
	// `end_cursor` (not the query's `end_cursor`).
	//
	// In a single transaction, subsequent query result batches for the same
	// query
	// can have a greater snapshot version number. Each batch's snapshot
	// version
	// is valid for all preceding batches.
	SnapshotVersion int64 `json:"snapshotVersion,omitempty,string"`
}

// ReadOptions: The options shared by read requests.
type ReadOptions struct {
	// ReadConsistency: The non-transactional read consistency to
	// use.
	// Cannot be set to `STRONG` for global queries.
	//
	// Possible values:
	//   "READ_CONSISTENCY_UNSPECIFIED" - Unspecified. This value must not
	// be used.
	//   "STRONG" - Strong consistency.
	//   "EVENTUAL" - Eventual consistency.
	ReadConsistency string `json:"readConsistency,omitempty"`

	// Transaction: The identifier of the transaction in which to read.
	// A
	// transaction identifier is returned by a call
	// to
	// Datastore.BeginTransaction.
	Transaction string `json:"transaction,omitempty"`
}

// RollbackRequest: The request for Datastore.Rollback.
type RollbackRequest struct {
	ProjectId string `json:"projectId"`
	// Transaction: The transaction identifier, returned by a call
	// to
	// Datastore.BeginTransaction.
	Transaction string `json:"transaction,omitempty"`
}

// RollbackResponse: The response for Datastore.Rollback.
// (an empty message).
type RollbackResponse struct {
	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// RunQueryRequest: The request for Datastore.RunQuery.
type RunQueryRequest struct {
	ProjectId string `json:"projectId"`
	// GqlQuery: The GQL query to run.
	GqlQuery *GqlQuery `json:"gqlQuery,omitempty"`

	// PartitionId: Entities are partitioned into subsets, identified by a
	// partition ID.
	// Queries are scoped to a single partition.
	// This partition ID is normalized with the standard default
	// context
	// partition ID.
	PartitionId *PartitionId `json:"partitionId,omitempty"`

	// Query: The query to run.
	Query *Query `json:"query,omitempty"`

	// ReadOptions: The options for this query.
	ReadOptions *ReadOptions `json:"readOptions,omitempty"`
}

// RunQueryResponse: The response for Datastore.RunQuery.
type RunQueryResponse struct {
	// Batch: A batch of query results (always present).
	Batch *QueryResultBatch `json:"batch,omitempty"`

	// Query: The parsed form of the `GqlQuery` from the request, if it was
	// set.
	Query *Query `json:"query,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	restful.ServerResponse `json:"-"`
}

// Value: A message that can hold any of the supported value types and
// associated
// metadata.
type Value struct {
	// ArrayValue: An array value.
	// Cannot contain another array value.
	// A `Value` instance that sets field `array_value` must not set
	// fields
	// `meaning` or `exclude_from_indexes`.
	ArrayValue *ArrayValue `json:"arrayValue,omitempty"`

	// BlobValue: A blob value.
	// May have at most 1,000,000 bytes.
	// When `exclude_from_indexes` is false, may have at most 1500 bytes.
	// In JSON requests, must be base64-encoded.
	BlobValue string `json:"blobValue,omitempty"`

	// BooleanValue: A boolean value.
	BooleanValue bool `json:"booleanValue,omitempty"`

	// DoubleValue: A double value.
	DoubleValue float64 `json:"doubleValue,omitempty"`

	// EntityValue: An entity value.
	//
	// - May have no key.
	// - May have a key with an incomplete key path.
	// - May have a reserved/read-only key.
	EntityValue *Entity `json:"entityValue,omitempty"`

	// ExcludeFromIndexes: If the value should be excluded from all indexes
	// including those defined
	// explicitly.
	ExcludeFromIndexes bool `json:"excludeFromIndexes,omitempty"`

	// GeoPointValue: A geo point value representing a point on the surface
	// of Earth.
	GeoPointValue *LatLng `json:"geoPointValue,omitempty"`

	// IntegerValue: An integer value.
	IntegerValue int64 `json:"integerValue,omitempty,string"`

	// KeyValue: A key value.
	KeyValue *Key `json:"keyValue,omitempty"`

	// Meaning: The `meaning` field should only be populated for backwards
	// compatibility.
	Meaning int64 `json:"meaning,omitempty"`

	// NullValue: A null value.
	//
	// Possible values:
	//   "NULL_VALUE" - Null value.
	NullValue string `json:"nullValue,omitempty"`

	// StringValue: A UTF-8 encoded string value.
	// When `exclude_from_indexes` is false (it is indexed) , may have at
	// most 1500 bytes.
	// Otherwise, may be set to at least 1,000,000 bytes.
	StringValue string `json:"stringValue,omitempty"`

	// TimestampValue: A timestamp value.
	// When stored in the Datastore, precise only to microseconds;
	// any additional precision is rounded down.
	TimestampValue string `json:"timestampValue,omitempty"`
}

type ProjectsService interface {
	AllocateIds(ctx *restful.Context, req *AllocateIdsRequest) (resp *AllocateIdsResponse, err error)
	BeginTransaction(ctx *restful.Context, req *BeginTransactionRequest) (resp *BeginTransactionResponse, err error)
	Commit(ctx *restful.Context, req *CommitRequest) (resp *CommitResponse, err error)
	Lookup(ctx *restful.Context, req *LookupRequest) (resp *LookupResponse, err error)
	Rollback(ctx *restful.Context, req *RollbackRequest) (resp *RollbackResponse, err error)
	RunQuery(ctx *restful.Context, req *RunQueryRequest) (resp *RunQueryResponse, err error)
}

type DefaultProjectsService struct {
}

func (s *DefaultProjectsService) AllocateIds(ctx *restful.Context, req *AllocateIdsRequest) (resp *AllocateIdsResponse, err error) {
	return nil, nil
}

func (s *DefaultProjectsService) BeginTransaction(ctx *restful.Context, req *BeginTransactionRequest) (resp *BeginTransactionResponse, err error) {
	return nil, nil
}

func (s *DefaultProjectsService) Commit(ctx *restful.Context, req *CommitRequest) (resp *CommitResponse, err error) {
	return nil, nil
}

func (s *DefaultProjectsService) Lookup(ctx *restful.Context, req *LookupRequest) (resp *LookupResponse, err error) {
	return nil, nil
}

func (s *DefaultProjectsService) Rollback(ctx *restful.Context, req *RollbackRequest) (resp *RollbackResponse, err error) {
	return nil, nil
}

func (s *DefaultProjectsService) RunQuery(ctx *restful.Context, req *RunQueryRequest) (resp *RunQueryResponse, err error) {
	return nil, nil
}

func RouteProjectsService(router restful.Router, service ProjectsService) (err error) {
	router.Handle("POST", "v1/projects/{projectId}:allocateIds", func(ctx *restful.Context) {
		req := &AllocateIdsRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.AllocateIds(ctx, req)
	})

	router.Handle("POST", "v1/projects/{projectId}:beginTransaction", func(ctx *restful.Context) {
		req := &BeginTransactionRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.BeginTransaction(ctx, req)
	})

	router.Handle("POST", "v1/projects/{projectId}:commit", func(ctx *restful.Context) {
		req := &CommitRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.Commit(ctx, req)
	})

	router.Handle("POST", "v1/projects/{projectId}:lookup", func(ctx *restful.Context) {
		req := &LookupRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.Lookup(ctx, req)
	})

	router.Handle("POST", "v1/projects/{projectId}:rollback", func(ctx *restful.Context) {
		req := &RollbackRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.Rollback(ctx, req)
	})

	router.Handle("POST", "v1/projects/{projectId}:runQuery", func(ctx *restful.Context) {
		req := &RunQueryRequest{}
		req.ProjectId = ctx.PathParamMap["projectId"]
		service.RunQuery(ctx, req)
	})

	return nil
}
