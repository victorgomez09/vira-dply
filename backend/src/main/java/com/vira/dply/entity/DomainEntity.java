package com.vira.dply.entity;

import com.vira.dply.enums.GatewayType;

import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "domains")
@Data
@AllArgsConstructor
@NoArgsConstructor
public class DomainEntity {
    @Id @GeneratedValue
    private Long id;

    private String host; // ej: myapp.example.com
    private Integer port; // puerto interno de la app, ej: 8080
    private Boolean tls;  // si se debe usar TLS

    @Enumerated(EnumType.STRING)
    private GatewayType gatewayType;

    @ManyToOne
    private ApplicationEntity application;
}