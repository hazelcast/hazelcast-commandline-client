package com.example;

import com.hazelcast.core.Hazelcast;
import com.hazelcast.core.HazelcastInstance;
import com.hazelcast.jet.config.JetConfig;
import com.hazelcast.jet.config.JobConfig;
import com.hazelcast.jet.pipeline.Pipeline;
import com.hazelcast.jet.pipeline.Sinks;
import com.hazelcast.jet.pipeline.test.TestSources;
import org.hashids.Hashids;

import java.util.AbstractMap;

public class SimplePipeline {
    public static void main(String[] args) {
        final String mapName = (args.length > 0)? args[0] : "my-map";
        // Create a pipeline that creates one item per second
        // and writes it to a map with the name given as an argument or my-map.
        Pipeline pipeline = Pipeline.create();
        pipeline.readFrom(TestSources.itemStream(1))
                .withoutTimestamps()
                .map(e -> new AbstractMap.SimpleEntry<>(
                        e.sequence(),
                        new Hashids().encode(e.sequence()))
                )
                .writeTo(Sinks.map(mapName,
                        // Map key is the sequence number
                        AbstractMap.SimpleEntry::getKey,
                        // Map value is the hash ID of the sequence
                        AbstractMap.SimpleEntry::getValue));
        HazelcastInstance hz = Hazelcast.bootstrappedInstance();
        JobConfig config = new JobConfig();
        // optionally set the name of the job
        config.setName("simple-pipeline");
        hz.getJet().newJob(pipeline, config);
    }
}
