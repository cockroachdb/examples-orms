package com.cockroachlabs.dialect;

import org.hibernate.dialect.PostgreSQL94Dialect;

public class CockroachDialect extends PostgreSQL94Dialect {

    /**
     * Override needed to avoid referencing "pg_class".
     * <p/>
     * {@inheritDoc}
     */
    @Override
    public String getQuerySequencesString() {
        return null;
    }

}
