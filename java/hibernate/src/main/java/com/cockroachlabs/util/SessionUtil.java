package com.cockroachlabs.util;

import com.cockroachlabs.model.Customer;
import com.cockroachlabs.model.Order;
import com.cockroachlabs.model.Product;
import org.hibernate.Session;
import org.hibernate.SessionFactory;
import org.hibernate.boot.registry.StandardServiceRegistryBuilder;
import org.hibernate.cfg.Configuration;
import org.hibernate.service.ServiceRegistry;

import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class SessionUtil {

    private static SessionUtil instance;

    private SessionFactory sessionFactory;

    private SessionUtil(String dbAddr) {
        if (dbAddr != null) {
            // When testing, anything output to stderr will fail the test, so we only
            // log severe messages that indicate serious failure.
            Logger.getLogger("org.hibernate").setLevel(Level.SEVERE);
        }

        Configuration configuration = new Configuration();
        configuration.configure("hibernate.cfg.xml");
        configuration.addAnnotatedClass(Customer.class);
        configuration.addAnnotatedClass(Order.class);
        configuration.addAnnotatedClass(Product.class);

        if (dbAddr != null) {
            // Most drivers expect the user in a connection to be specified like:
            //   postgresql://<user>@host:port/db
            // but the PGJDBC driver expects the user as a parameter like:
            //   postgresql://host:port/db
            // with the username and password passed as separate properties.
            Pattern p = Pattern.compile("postgresql://((\\w+)(:(\\w+))?@).*");
            Matcher m = p.matcher(dbAddr);
            if (m.matches()) {
                String userPart = m.group(1);
                String user = m.group(2);
                String password = m.group(4);

                dbAddr = dbAddr.replace(userPart, "");
                configuration.setProperty("hibernate.connection.user", user);
                if (password != null && !password.equals("")) {
                    configuration.setProperty("hibernate.connection.password", password);
                }
            }

            // The client cert must be in PKCS8 format for Java.
            dbAddr = dbAddr.replace("client.root.key", "client.root.key.pk8");

            // Add the "jdbc:" prefix to the address and replace in configuration.
            dbAddr = "jdbc:" + dbAddr;
            configuration.setProperty("hibernate.connection.url", dbAddr);
        }

        ServiceRegistry serviceRegistry = new StandardServiceRegistryBuilder()
                .applySettings(configuration.getProperties())
                .build();

        sessionFactory = configuration.buildSessionFactory(serviceRegistry);
    }

    public static void init(String dbAddr) {
        instance = new SessionUtil(dbAddr);
    }

    private static SessionUtil getInstance() {
        return instance;
    }

    public static Session getSession(){
        return getInstance().sessionFactory.openSession();
    }

}
