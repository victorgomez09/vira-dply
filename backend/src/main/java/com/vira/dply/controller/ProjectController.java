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

import com.vira.dply.dto.ProjectDto;
import com.vira.dply.mapper.ProjectMapper;
import com.vira.dply.security.JwtUserPrincipal;
import com.vira.dply.service.ProjectService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/projects")
@RequiredArgsConstructor
public class ProjectController {

    private final ProjectService projectService;
    private final ProjectMapper projectMapper;

    @GetMapping("/environments/{envId}")
    public ResponseEntity<List<ProjectDto>> getProjects(
            @PathVariable UUID envId,
            Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.OK).body(
                projectService.findByEnvironment(envId, user.userId()).stream().map(projectMapper::toDto).toList());
    }

    @PostMapping("/environments/{envId}")
    @ResponseStatus(HttpStatus.CREATED)
    public ResponseEntity<ProjectDto> createProject(
            @PathVariable UUID envId,
            @RequestBody ProjectDto request,
            Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.CREATED).body(projectMapper
                .toDto(projectService.createProject(envId, user.userId(), projectMapper.toEntity(request))));
    }

    @GetMapping("/projects/{projectId}")
    public ResponseEntity<ProjectDto> getProject(
            @PathVariable UUID projectId, Authentication authentication) {
        JwtUserPrincipal user = (JwtUserPrincipal) authentication.getPrincipal();

        return ResponseEntity.status(HttpStatus.ACCEPTED)
                .body(projectMapper.toDto(projectService.findById(projectId, user.userId())));
    }
}