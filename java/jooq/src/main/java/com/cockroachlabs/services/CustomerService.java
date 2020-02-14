package com.cockroachlabs.services;

import com.cockroachlabs.example.jooq.db.tables.pojos.Customers;
import com.cockroachlabs.example.jooq.db.tables.records.CustomersRecord;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import javax.ws.rs.*;
import java.io.IOException;
import java.util.List;

import static com.cockroachlabs.example.jooq.db.Tables.*;
import static com.cockroachlabs.util.Utils.ctx;
import static org.jooq.impl.DSL.select;

@Path("/customer")
public class CustomerService {

    private final ObjectMapper mapper = new ObjectMapper();

    @GET
    @Produces("application/json")
    public String getCustomers() {
        try {
            List<Customers> customers = ctx()
                .selectFrom(CUSTOMERS)
                .fetchInto(Customers.class);

            return mapper.writeValueAsString(customers);
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @POST
    @Produces("application/json")
    public String createCustomer(String body) {
        try {
            CustomersRecord record = ctx().newRecord(CUSTOMERS, mapper.readValue(body, Customers.class));
            ctx().executeInsert(record);

            return mapper.writeValueAsString(record.into(Customers.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @GET
    @Path("/{customerID}")
    @Produces("application/json")
    public String getCustomer(@PathParam("customerID") long customerID) {
        try {
            CustomersRecord record = ctx().fetchOne(CUSTOMERS, CUSTOMERS.ID.eq(customerID));
            if (record == null)
                throw new NotFoundException();

            return mapper.writeValueAsString(record.into(Customers.class));
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @PUT
    @Path("/{customerID}")
    @Produces("application/json")
    public String updateCustomer(@PathParam("customerID") long customerID, String body) {
        try {
            CustomersRecord record = ctx().newRecord(CUSTOMERS, mapper.readValue(body, Customers.class));
            record.setId(customerID);
            record.reset(CUSTOMERS.ID);
            record.update();

            return mapper.writeValueAsString(record.into(Customers.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @DELETE
    @Path("/{customerID}")
    @Produces("text/plain")
    public String deleteCustomer(@PathParam("customerID") long customerID) {
        return ctx().transactionResult(ctx -> {
            ctx.dsl()
                .deleteFrom(ORDER_PRODUCTS)
                .where(ORDER_PRODUCTS.ORDER_ID.in(
                    select(ORDERS.ID)
                    .from(ORDERS)
                    .where(ORDERS.CUSTOMER_ID.eq(customerID))
                ))
                .execute();

            ctx.dsl()
                .deleteFrom(ORDERS)
                .where(ORDERS.CUSTOMER_ID.eq(customerID))
                .execute();

            int rowCount = ctx.dsl()
                .deleteFrom(CUSTOMERS)
                .where(CUSTOMERS.ID.eq(customerID))
                .execute();

            if (rowCount == 0)
                throw new NotFoundException();

            return "ok";
        });
    }

}
