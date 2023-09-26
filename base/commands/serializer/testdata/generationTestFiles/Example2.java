package com.example;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class Example2 {

    public static final class Serializer implements CompactSerializer<Example2> {
        @Nonnull
        @Override
        public Example2 read(@Nonnull CompactReader reader) {
            int foo = reader.readInt32("foo");
            return new Example2(foo);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull Example2 object) {
            writer.writeInt32("foo", object.foo);
        }

        @Override
        public Class<Example2> getCompactClass() {
            return Example2.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<Example2> HZ_COMPACT_SERIALIZER = new Serializer();

    private int foo;

    public Example2() {
    }

    public Example2(int foo) {
        this.foo = foo;
    }

    public int getFoo() {
        return foo;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Example2 that = (Example2) o;
        if (foo != that.foo) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (int) foo;

        return result;
    }

    @Override
    public String toString() {
        return "<Example2> {"
                + "foo=" + foo
                + '}';
    }

}