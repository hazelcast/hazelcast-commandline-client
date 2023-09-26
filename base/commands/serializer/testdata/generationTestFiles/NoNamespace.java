import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class NoNamespace {

    public static final class Serializer implements CompactSerializer<NoNamespace> {
        @Nonnull
        @Override
        public NoNamespace read(@Nonnull CompactReader reader) {
            boolean foo = reader.readBoolean("foo");
            NoNamespaceNested bar = reader.readCompact("bar");
            return new NoNamespace(foo, bar);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull NoNamespace object) {
            writer.writeBoolean("foo", object.foo);
            writer.writeCompact("bar", object.bar);
        }

        @Override
        public Class<NoNamespace> getCompactClass() {
            return NoNamespace.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<NoNamespace> HZ_COMPACT_SERIALIZER = new Serializer();

    private boolean foo;
    private NoNamespaceNested bar;

    public NoNamespace() {
    }

    public NoNamespace(boolean foo, NoNamespaceNested bar) {
        this.foo = foo;
        this.bar = bar;
    }

    public boolean getFoo() {
        return foo;
    }

    public NoNamespaceNested getBar() {
        return bar;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        NoNamespace that = (NoNamespace) o;
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
        return "<NoNamespace> {"
                + "foo=" + foo
                + ", bar=" + bar
                + '}';
    }

}