An in-memory key-value data store using a custom simple binary protocol.


Takes advantage of Go's concurrency capabilities for handling mulitple clients conccurently by using a 2 goroutines per client model. One for reading and one for writing.
This may seem expensive at first glance, but goroutines are incredibally cheap and although they can dynamically become more and more expensive, most of the work of a handler is de-serialization and serialization and io waiting. And using a binary protocol instead can help make this little cpu bound work even more efficient.


Also implements sharded data store model where we partition the map into N<=Cores and let each core handle a subset of the data. This greatly helps improve performance because of better cache locality. Partitioning the data and letting each core handle its own shard greatly helps in reducing unneccesary cache invalidation by eliminating any cross-core communication. Aligning the shard according to cache line sizes will probably eliminate this fully.


The executor communicates with all the client handlers using a queue-based mechanism (go's buffered channels) to prevent any race conditions and lock-based access. 
Each shard gets it's own request queue and requests will be routed accordingly using a hash function. This is basically how a normal single-threaded architecture(Redis) would work but instead of one executor handling all requests and handling the full map, we divide the map and the work. The response from the executor gets communicated via a private channel reserved for the client. In other words, channel-per-client.


Essentially, we have a Redis instance per core model but each core handles only a subset of the map.
This way we get the concurrency and race-free mechanism of a single-threaded model along with the power of parallelism of a multi-threaded model, with the bonus of performance boost because of cache locality and reduced cache invalidations.


DragonflyDB is one of the in-memory databases which uses this shared-nothing architecture. It is a rewrite of Redis(uses RESP protocol) using c++ which aims to use the power of multi-threading to improve the performance, and make the most of, a single instance instead of approching distributed model for scaling.
