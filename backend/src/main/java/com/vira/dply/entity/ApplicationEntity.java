package com.vira.dply.entity;

import java.time.Instant;
import java.util.UUID;

import org.hibernate.annotations.CreationTimestamp;

import com.vira.dply.enums.ApplicationStatus;
import com.vira.dply.enums.ApplicationType;
import com.vira.dply.enums.BuildStatus;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "applications", uniqueConstraints = {
        @UniqueConstraint(columnNames = { "project_id", "name" })
})
@Data
@AllArgsConstructor
@NoArgsConstructor
public class ApplicationEntity {

    @Id
    @GeneratedValue
    private UUID id;

    @Column(nullable = false)
    private String name;

    private String description;

    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "project_id", nullable = false)
    private ProjectEntity project;

    @Column(nullable = false)
    private String gitRepository;

    private String gitBranch = "main";

    private String gitUsername;
    
    private String gitPasswordOrToken;
    
    private String gitPrivateKey;
    
    private String gitPassphrase;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ApplicationType type;

    @Enumerated(EnumType.STRING)
    private BuildStatus buildStatus;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ApplicationStatus status = ApplicationStatus.CREATED;

    @Column(nullable = false)
    private int replicas = 1;

    private String imageName;
    private String buildLogs;

    private boolean autoScalable = false;

    @CreationTimestamp
    private Instant createdAt = Instant.now();
}