package com.vira.dply.controller;

import java.util.List;
import java.util.UUID;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import com.vira.dply.dto.ApplicationDto;
import com.vira.dply.mapper.ApplicationMapper;
import com.vira.dply.security.JwtUserPrincipal;
import com.vira.dply.service.ApplicationService;
import com.vira.dply.service.AsyncDeploymentProcessor;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/applications")
@RequiredArgsConstructor
public class ApplicationController {

    private final ApplicationService applicationService;
    private final AsyncDeploymentProcessor asyncDeploymentProcessor;
    private final ApplicationMapper applicationMapper;

    @GetMapping("/{projectId}/{applicationId}")
    public ResponseEntity<ApplicationDto> findByID(@PathVariable("projectId") UUID projectId,
            @PathVariable("applicationId") UUID applicationId, Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.OK)
                .body(applicationMapper.toDto(applicationService.findById(projectId, projectId, user.userId())));
    }

    @GetMapping("/{projectId}")
    public ResponseEntity<List<ApplicationDto>> findByProjectID(@PathVariable("projectId") UUID projectId,
            Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.OK)
                .body(applicationService.findByProjectyId(projectId, user.userId()).stream()
                        .map(applicationMapper::toDto).toList());
    }

    @PostMapping("/projects/{projectId}/applications")
    @ResponseStatus(HttpStatus.CREATED)
    public ResponseEntity<ApplicationDto> create(
            @PathVariable UUID projectId,
            @RequestBody ApplicationDto request,
            Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.CREATED).body(applicationMapper
                .toDto(applicationService.create(projectId, user.userId(), applicationMapper.toEntity(request))));
    }

    @PostMapping("/applications/{appId}/deploy")
    @ResponseStatus(HttpStatus.ACCEPTED)
    public void deploy(@PathVariable UUID appId, Authentication authentication) {
        asyncDeploymentProcessor.buildAndDeployApplication(appId);
    }
}
