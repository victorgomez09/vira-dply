package com.vira.dply.service

import com.vira.dply.model.EnvironmentStatus
import com.vira.dply.model.Team
import com.vira.dply.model.TeamStatus
import com.vira.dply.repository.EnvironmentRepository
import com.vira.dply.repository.TeamRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import org.springframework.stereotype.Service

@Service
class TeamService(
    private val teamRepository: TeamRepository,
    private val envRepository: EnvironmentRepository,
    private val namespaceService: NamespaceService,
    private val rbacService: RbacService
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    fun createTeam(
        envId: Long,
        teamName: String
    ): Team {
        val env = envRepository.findById(envId)
            .orElseThrow { error("Environment not found") }

        require(env.status == EnvironmentStatus.READY) {
            "Environment is not ready"
        }

        val team = Team(
            id = 0L,
            name = teamName,
            environment = env
        )

        val saved = teamRepository.save(team)

        scope.launch {
            try {
                namespaceService.createNamespace(env, teamName)
                rbacService.setupTeamRbac(env, teamName)
                saved.status = TeamStatus.READY
            } catch (ex: Exception) {
                saved.status = TeamStatus.FAILED
            }
            teamRepository.save(saved)
        }

        return saved
    }
}