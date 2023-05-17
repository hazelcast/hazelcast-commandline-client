package hz.primes

import com.hazelcast.core.Hazelcast
import com.hazelcast.jet.pipeline.Pipeline
import com.hazelcast.jet.pipeline.Sinks
import com.hazelcast.jet.pipeline.test.TestSources


fun main() {
    val pipeline = Pipeline.create()
    pipeline.readFrom(TestSources.itemStream(1) )
        .withoutTimestamps()
        .filter { isPrime(it.sequence()) }
        .setName("write prime numbers into map with their timestamps")
        .writeTo(Sinks.map(
            "primes",
            { it.sequence() },
            { it.timestamp() },
         ))
    val hz = Hazelcast.bootstrappedInstance()
    hz.jet.newJob(pipeline)
}

fun isPrime(num: Long): Boolean {
    if (num < 2) {
        return false
    }
    for (i in 2 until num) {
        if (num.mod(i) == 0L) {
            return false
        }
    }
    return true
}
