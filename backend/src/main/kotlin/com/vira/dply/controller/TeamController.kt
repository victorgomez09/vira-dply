package com.vira.dply.controller

import com.vira.dply.dto.TeamDto
import com.vira.dply.mapper.TeamMapper
import com.vira.dply.service.TeamService
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/environments/{envId}/teams")
class TeamController(
    private val teamService: TeamService,
    private val teamMapper: TeamMapper
) {

    @PostMapping
    suspend fun create(
        @PathVariable envId: Long,
        @RequestBody payload: String
    ): ResponseEntity<TeamDto> {
        val team = teamService.createTeam(envId, payload)

        return ResponseEntity.ok(teamMapper.toDto(team))
    }
}