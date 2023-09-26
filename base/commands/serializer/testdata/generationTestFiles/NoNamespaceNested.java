import com.hazelcast.nio.serialization.compact.CompactReader;
import com.hazelcast.nio.serialization.compact.CompactSerializer;
import com.hazelcast.nio.serialization.compact.CompactWriter;

import javax.annotation.Nonnull;
import java.util.Arrays;
import java.util.Objects;

public class NoNamespaceNested {

    public static final class Serializer implements CompactSerializer<NoNamespaceNested> {
        @Nonnull
        @Override
        public NoNamespaceNested read(@Nonnull CompactReader reader) {
            boolean baz = reader.readBoolean("baz");
            return new NoNamespaceNested(baz);
        }

        @Override
        public void write(@Nonnull CompactWriter writer, @Nonnull NoNamespaceNested object) {
            writer.writeBoolean("baz", object.baz);
        }

        @Override
        public Class<NoNamespaceNested> getCompactClass() {
            return NoNamespaceNested.class;
        }

        @Override
        public String getTypeName() {
            return "00000000-0000-000a-0000-00000000000a";
        }
    };

    public static final CompactSerializer<NoNamespaceNested> HZ_COMPACT_SERIALIZER = new Serializer();

    private boolean baz;

    public NoNamespaceNested() {
    }

    public NoNamespaceNested(boolean baz) {
        this.baz = baz;
    }

    public boolean getBaz() {
        return baz;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        NoNamespaceNested that = (NoNamespaceNested) o;
        if (baz != that.baz) return false;

        return true;
    }

    @Override
    public int hashCode() {
        int result = 0;
        result = 31 * result + (baz ? 1 : 0);

        return result;
    }

    @Override
    public String toString() {
        return "<NoNamespaceNested> {"
                + "baz=" + baz
                + '}';
    }

}