package com.cockroachlabs;

import com.cockroachlabs.model.Customer;
import com.cockroachlabs.model.Order;
import com.cockroachlabs.model.Product;
import org.hibernate.Session;
import org.hibernate.SessionFactory;
import org.hibernate.boot.registry.StandardServiceRegistryBuilder;
import org.hibernate.cfg.Configuration;
import org.hibernate.service.ServiceRegistry;

import java.math.BigDecimal;

public class Application {

    public static void main(String[] args) {
        try (SessionFactory sf = buildSessionFactory()) {
            Session session = sf.getCurrentSession();

            Customer c = new Customer();
            c.setName("joe");

            Order o = new Order();
            o.setCustomer(c);
            o.setSubtotal(new BigDecimal(100));

            session.beginTransaction();
            session.save(c);
            session.save(o);
            session.getTransaction().commit();
        }
    }

    private static SessionFactory buildSessionFactory() {
        Configuration configuration = new Configuration();
        configuration.configure("hibernate.cfg.xml");
        configuration.addAnnotatedClass(Customer.class);
        configuration.addAnnotatedClass(Order.class);
        configuration.addAnnotatedClass(Product.class);

        ServiceRegistry serviceRegistry = new StandardServiceRegistryBuilder()
                .applySettings(configuration.getProperties())
                .build();

        return configuration.buildSessionFactory(serviceRegistry);
    }

}
