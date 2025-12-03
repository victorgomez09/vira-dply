package com.vira.dply.controller;

import java.util.List;
import java.util.UUID;

import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.vira.dply.entity.TeamEntity;
import com.vira.dply.service.TeamService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/teams")
@RequiredArgsConstructor
public class TeamController {

    private final TeamService teamService;

    @PostMapping
    public TeamEntity createTeam(@RequestParam String name,
                                 @RequestParam UUID environmentId) {
        return teamService.createTeam(name, environmentId);
    }

    @GetMapping
    public List<TeamEntity> listTeams(@RequestParam UUID environmentId) {
        return teamService.listTeams(environmentId);
    }

    @DeleteMapping("/{id}")
    public void deleteTeam(@PathVariable UUID id) {
        teamService.deleteTeam(id);
    }
}