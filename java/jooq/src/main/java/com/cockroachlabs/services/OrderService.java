package com.cockroachlabs.services;

import com.cockroachlabs.example.jooq.db.tables.pojos.Orders;
import com.cockroachlabs.example.jooq.db.tables.records.OrdersRecord;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import javax.ws.rs.*;
import java.io.IOException;
import java.util.List;

import static com.cockroachlabs.example.jooq.db.Tables.ORDERS;
import static com.cockroachlabs.example.jooq.db.Tables.ORDER_PRODUCTS;
import static com.cockroachlabs.util.Utils.ctx;

@Path("/order")
public class OrderService {

    private final ObjectMapper mapper = new ObjectMapper();

    @GET
    @Produces("application/json")
    public String getOrders() {
        try {
            List<Orders> orders = ctx()
                .selectFrom(ORDERS)
                .fetchInto(Orders.class);

            return mapper.writeValueAsString(orders);
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @POST
    @Produces("application/json")
    public String createOrder(String body) {
        try {
            OrdersRecord record = ctx().newRecord(ORDERS, mapper.readValue(body, Orders.class));
            ctx().executeInsert(record);

            return mapper.writeValueAsString(record.into(Orders.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @GET
    @Path("/{orderID}")
    @Produces("application/json")
    public String getOrder(@PathParam("orderID") long orderID) {
        try {
            OrdersRecord record = ctx().fetchOne(ORDERS, ORDERS.ID.eq(orderID));
            if (record == null)
                throw new NotFoundException();

            return mapper.writeValueAsString(record.into(Orders.class));
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @PUT
    @Path("/{orderID}")
    @Produces("application/json")
    public String updateOrder(@PathParam("orderID") long orderID, String body) {
        try {
            OrdersRecord record = ctx().newRecord(ORDERS, mapper.readValue(body, Orders.class));
            record.setId(orderID);
            record.reset(ORDERS.ID);
            record.update();

            return mapper.writeValueAsString(record.into(Orders.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @DELETE
    @Path("/{orderID}")
    @Produces("text/plain")
    public String deleteOrder(@PathParam("orderID") long orderID) {
        return ctx().transactionResult(ctx -> {
            ctx.dsl()
                .deleteFrom(ORDER_PRODUCTS)
                .where(ORDER_PRODUCTS.ORDER_ID.eq(orderID))
                .execute();

            int rowCount = ctx.dsl()
                .deleteFrom(ORDERS)
                .where(ORDERS.ID.eq(orderID))
                .execute();

            if (rowCount == 0)
                throw new NotFoundException();

            return "ok";
        });
    }

}
