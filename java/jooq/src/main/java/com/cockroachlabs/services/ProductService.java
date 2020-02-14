package com.cockroachlabs.services;

import com.cockroachlabs.example.jooq.db.tables.pojos.Customers;
import com.cockroachlabs.example.jooq.db.tables.pojos.Products;
import com.cockroachlabs.example.jooq.db.tables.records.ProductsRecord;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.jooq.exception.DataAccessException;

import javax.ws.rs.*;
import java.io.IOException;
import java.util.List;

import static com.cockroachlabs.example.jooq.db.Tables.ORDER_PRODUCTS;
import static com.cockroachlabs.example.jooq.db.Tables.PRODUCTS;
import static com.cockroachlabs.util.Utils.ctx;

@Path("/product")
public class ProductService {

    private final ObjectMapper mapper = new ObjectMapper();

    @GET
    @Produces("application/json")
    public String getProducts() {
        try {
            List<Products> customers = ctx()
                .selectFrom(PRODUCTS)
                .fetchInto(Products.class);

            return mapper.writeValueAsString(customers);
        } catch (DataAccessException e) {
            e.printStackTrace();
            throw e;
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @POST
    @Produces("application/json")
    public String createProduct(String body) {
        try {
            ProductsRecord record = ctx().newRecord(PRODUCTS, mapper.readValue(body, Products.class));
            ctx().executeInsert(record);

            return mapper.writeValueAsString(record.into(Products.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @GET
    @Path("/{productID}")
    @Produces("application/json")
    public String getProduct(@PathParam("productID") long productID) {
        try {
            ProductsRecord record = ctx().fetchOne(PRODUCTS, PRODUCTS.ID.eq(productID));
            if (record == null)
                throw new NotFoundException();

            return mapper.writeValueAsString(record.into(Customers.class));
        } catch (JsonProcessingException e) {
            return e.toString();
        }
    }

    @PUT
    @Path("/{productID}")
    @Produces("application/json")
    public String updateProduct(@PathParam("productID") long productID, String body) {
        try {
            ProductsRecord record = ctx().newRecord(PRODUCTS, mapper.readValue(body, Products.class));
            record.setId(productID);
            record.reset(PRODUCTS.ID);
            record.update();

            return mapper.writeValueAsString(record.into(Products.class));
        } catch (IOException e) {
            return e.toString();
        }
    }

    @DELETE
    @Path("/{productID}")
    @Produces("text/plain")
    public String deleteProduct(@PathParam("productID") long productID) {
        return ctx().transactionResult(ctx -> {
             ctx.dsl()
                .deleteFrom(ORDER_PRODUCTS)
                .where(ORDER_PRODUCTS.PRODUCT_ID.eq(productID))
                .execute();

            int rowCount = ctx.dsl()
                .deleteFrom(PRODUCTS)
                .where(PRODUCTS.ID.eq(productID))
                .execute();

            if (rowCount == 0)
                throw new NotFoundException();

            return "ok";
        });
    }

}
