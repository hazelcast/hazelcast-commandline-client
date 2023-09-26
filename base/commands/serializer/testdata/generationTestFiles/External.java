import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import com.x.y.z.SomeExternalClass;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class External {

    public static final class Serializer implements CompactSerializer<External> {
        @Nonnull
        @Override
        public External read(@Nonnull CompactReader reader) {
            boolean foo = reader.readBoolean("foo");
            com.x.y.z.SomeExternalClass bar = reader.readCompact("bar");
            return new External(foo, bar);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull External object) {
            writer.writeBoolean("foo", object.foo);
            writer.writeCompact("bar", object.bar);
        }

        @Override
        public Class<External> getCompactClass() {
            return External.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<External> HZ_COMPACT_SERIALIZER = new Serializer();

    private boolean foo;
    private com.x.y.z.SomeExternalClass bar;

    public External() {
    }

    public External(boolean foo, com.x.y.z.SomeExternalClass bar) {
        this.foo = foo;
        this.bar = bar;
    }

    public boolean getFoo() {
        return foo;
    }

    public com.x.y.z.SomeExternalClass getBar() {
        return bar;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        External that = (External) o;
        if (foo != that.foo) return false;
        if (!Objects.equals(bar, that.bar)) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (foo ? 1 : 0);
        result = 31 * result + Objects.hashCode(bar);

        return result;
    }

    @Override
    public String toString() {
        return "<External> {"
                + "foo=" + foo
                + ", bar=" + bar
                + '}';
    }

}