package com.vira.dply.util

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.eclipse.jgit.api.Git
import org.eclipse.jgit.transport.UsernamePasswordCredentialsProvider
import org.springframework.stereotype.Component
import java.nio.file.Path

@Component
class JGitRepositoryCloner {
    suspend fun clone(
        source: GitSource,
        destination: Path,
        subPath: String?
    ): GitCloneResult = withContext(Dispatchers.IO) {
        try {
            val cloneCommand = Git.cloneRepository()
                .setURI(source.url)
                .setDirectory(destination.toFile())
                .setDepth(1)
                .setBranch(source.ref)

            if (!source.authToken.isNullOrBlank()) {
                cloneCommand.setCredentialsProvider(
                    UsernamePasswordCredentialsProvider(source.authToken, "")
                )
            }

            val git = cloneCommand.call()

            // sparse checkout if subPath present
            if (!subPath.isNullOrBlank() && subPath != ".") {
                configureSparseCheckout(git, subPath)
            }

            val head = git.repository.resolve("HEAD").name

            GitCloneResult(
                commitSha = head.take(12)
            )
        } catch (ex: GitAPIException) {
            throw IllegalStateException("Git clone failed", ex)
        }
    }

    private fun configureSparseCheckout(git: Git, subPath: String) {
        val repo = git.repository
        val sparseFile = repo.directory.resolve("info/sparse-checkout")

        sparseFile.parentFile.mkdirs()
        sparseFile.writeText("$subPath/*\n")

        repo.config.apply {
            setBoolean("core", null, "sparseCheckout", true)
            save()
        }

        git.checkout().setName("HEAD").call()
    }
}