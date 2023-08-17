package com.example;

import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class NestedCompact {

    public static final class Serializer implements CompactSerializer<NestedCompact> {
        @Nonnull
        @Override
        public NestedCompact read(@Nonnull CompactReader reader) {
            java.lang.String foo = reader.readString("foo");
            int bar = reader.readInt32("bar");
            return new NestedCompact(foo, bar);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull NestedCompact object) {
            writer.writeString("foo", object.foo);
            writer.writeInt32("bar", object.bar);
        }
    };

    public static final CompactSerializer<NestedCompact> HZ_COMPACT_SERIALIZER = new Serializer();

    private java.lang.String foo;
    private int bar;

    public NestedCompact() {
    }

    public NestedCompact(java.lang.String foo, int bar) {
        this.foo = foo;
        this.bar = bar;
    }

    public java.lang.String getFoo() {
        return foo;
    }

    public int getBar() {
        return bar;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        NestedCompact that = (NestedCompact) o;
        if (!Objects.equals(foo, that.foo)) return false;
        if (bar != that.bar) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + Objects.hashCode(foo);
        result = 31 * result + (int) bar;

        return result;
    }

    @Override
    public String toString() {
        return "<NestedCompact> {"
                + "foo=" + foo
                + ", bar=" + bar
                + '}';
    }

}