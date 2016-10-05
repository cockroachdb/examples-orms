package com.cockroachlabs.model;

import javax.persistence.*;
import java.math.BigDecimal;
import java.util.Set;

@Entity
@Table(name="products")
public class Product {

    @Id
    @GeneratedValue(strategy=GenerationType.IDENTITY)
    @Column(name="ID", nullable=false, unique=true)
    private long id;

    @Column(name="NAME")
    private String name;

    @Column(name="PRODUCT", precision=18, scale=2)
    private BigDecimal product;

    @ManyToMany(cascade=CascadeType.ALL, mappedBy="products")
    private Set<Order> orders;

    public long getId() {
        return id;
    }

    public void setId(long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public BigDecimal getProduct() {
        return product;
    }

    public void setProduct(BigDecimal product) {
        this.product = product;
    }

    public Set<Order> getOrders() {
        return orders;
    }

    public void setOrders(Set<Order> orders) {
        this.orders = orders;
    }

}
