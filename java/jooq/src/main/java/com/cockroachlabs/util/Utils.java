package com.cockroachlabs.util;

import com.zaxxer.hikari.HikariDataSource;
import org.jooq.DSLContext;
import org.jooq.SQLDialect;
import org.jooq.Source;
import org.jooq.conf.RenderQuotedNames;
import org.jooq.conf.Settings;
import org.jooq.impl.DSL;
import org.jooq.impl.DefaultConfiguration;

import java.io.InputStream;

public class Utils {

    private static DSLContext ctx;
    private static String dbAddr;

    public static DSLContext ctx() {
        if (ctx == null) {
            try {
                HikariDataSource ds = new HikariDataSource();
                ds.setJdbcUrl(dbAddr != null ? dbAddr : "jdbc:postgresql://localhost:26257/postgres");
                ds.setUsername("root");
                ds.setPassword("");

                ctx = DSL.using(ds, SQLDialect.COCKROACHDB,
                    new Settings()
                        .withExecuteLogging(true)
                        .withRenderQuotedNames(RenderQuotedNames.EXPLICIT_DEFAULT_UNQUOTED)
                );

                // Initialise schema
                try (InputStream in = Utils.class.getResourceAsStream("/db.sql")) {
                    ctx.parser().parse(Source.of(in).readString()).executeBatch();
                }

                Runtime.getRuntime().addShutdownHook(new Thread() {
                    @Override
                    public void run() {
                        ds.close();
                    }
                });
            }
            catch (Exception e) {
                throw new RuntimeException(e);
            }
        }

        return ctx;
    }

    public static void init(String dbAddr) {
        Utils.dbAddr = dbAddr;
        ctx();
    }
}
