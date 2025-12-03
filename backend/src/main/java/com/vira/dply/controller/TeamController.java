package com.vira.dply.controller;

import java.util.List;
import java.util.UUID;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.vira.dply.dto.TeamDto;
import com.vira.dply.dto.UserTeamDto;
import com.vira.dply.mapper.TeamMapper;
import com.vira.dply.security.JwtUserPrincipal;
import com.vira.dply.service.TeamService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/teams")
@RequiredArgsConstructor
public class TeamController {

    private final TeamService teamService;
    private final TeamMapper teamMapper;

    @PostMapping("/{environmentId}")
    public ResponseEntity<TeamDto> createTeam(@PathVariable("environmentId") UUID environmentId, @RequestBody String name, Authentication auth) {
        JwtUserPrincipal user = (JwtUserPrincipal) auth.getPrincipal();

        return ResponseEntity.status(HttpStatus.CREATED).body(teamMapper.toDto(
                teamService.createTeam(name, environmentId, user.userId())));
    }

    @PostMapping("/add-user")
    public ResponseEntity<TeamDto> addUserToTeam(@RequestBody UserTeamDto payload) {
        return ResponseEntity.status(HttpStatus.ACCEPTED).body(teamMapper
                .toDto(teamService.addUserToTeam(payload.getTeamId(), payload.getUserId(), payload.getRole())));
    }

    @PostMapping("/remove-user")
    public ResponseEntity<Void> updateUserRole(@RequestBody UserTeamDto payload) {
        teamService.removeUserFromTeam(payload.getTeamId(), payload.getUserId());

        return ResponseEntity.status(HttpStatus.ACCEPTED).build();
    }

    @PostMapping("/edit-role")
    public ResponseEntity<TeamDto> updateUserRole(@RequestBody UserTeamDto payload, Authentication auth) {
        return ResponseEntity.status(HttpStatus.ACCEPTED).body(teamMapper
                .toDto(teamService.updateUserRole(payload.getTeamId(), payload.getUserId(), payload.getRole())));
    }

    @GetMapping
    public ResponseEntity<List<TeamDto>> listTeams(@RequestParam UUID environmentId) {
        return ResponseEntity.status(HttpStatus.OK).body(teamService.listTeams(environmentId).stream().map(teamMapper::toDto).toList());
    }

    @DeleteMapping("/{id}")
    public void deleteTeam(@PathVariable UUID id) {
        teamService.deleteTeam(id);
    }
}