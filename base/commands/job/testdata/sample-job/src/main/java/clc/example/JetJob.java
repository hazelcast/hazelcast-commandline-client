package clc.example;

import com.hazelcast.core.Hazelcast;
import com.hazelcast.jet.pipeline.Pipeline;
import com.hazelcast.jet.pipeline.Sinks;
import com.hazelcast.jet.pipeline.test.TestSources;

import java.util.AbstractMap;

public class JetJob {
    public static void main(String[] args) {
        final var mapName = (args.length > 0)? args[0] : "my-map";
        var pipeline = Pipeline.create();
        pipeline.readFrom(TestSources.itemStream(1))
                .withoutTimestamps()
                .map(e -> new AbstractMap.SimpleEntry<>(e.sequence(), 10 * e.sequence()))
                .writeTo(Sinks.map(mapName,
                        AbstractMap.SimpleEntry::getKey,
                        AbstractMap.SimpleEntry::getValue));
        var hz = Hazelcast.bootstrappedInstance();
        hz.getJet().newJob(pipeline);
    }
}
