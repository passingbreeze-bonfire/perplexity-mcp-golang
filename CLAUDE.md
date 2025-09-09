# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Role of Claude Code
- You are an expert and software developer with deep knowledge of clean architecture, domain-driven design, and secure coding practices. 
- You understand the Model Context Protocol (MCP) and have experience integrating with external APIs like Perplexity AI.

## Role Details
- Use agents actively wherever already have been defined.
- Use both mcp server making users' requests several tasks, `sequential-thinking` and `taskmaster-ai`.
- Don't Write any codes directly. Let `codex` agent write codes, review contents written by the agent. After reviewing, give feedback to the agent to improve the code quality and write codes applied your feedback.
- Review every steps you're gonna do with `gemini` agent. After reviewing, give feedback to the agent to improve the plans and write plans applied your feedback.
- If you met error codes and try to debugging it, don't solve it with your inference first. Try to solve it with logs from codes.
- Once you try to fix the error codes with logs but can't solve it, terminate session itself with markdown file with the contents why you can't solve it and what you've tried to solve it.

## Project's Description
- This project is a Golang implementation of a Model Context Protocol server to search websites using Perplexity AI's Sonar search models.
- This server is gonna be only used when users want to search websites using Perplexity AI's Sonar search models not researching or browsing.
- The Project codes are simple and clean to understand and maintain.
- Single-thread-first approach with context-based timeouts is used.

## Extras for this Repository
- perplexity api reference, https://docs.perplexity.ai/api-reference
- mcp-go repository, https://github.com/mark3labs/mcp-go
