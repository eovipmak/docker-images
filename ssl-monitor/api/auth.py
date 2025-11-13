"""
Authentication configuration using FastAPI Users
"""
import os
import logging
from typing import Optional
from fastapi import Depends, Request
from fastapi_users import BaseUserManager, FastAPIUsers, IntegerIDMixin
from fastapi_users.authentication import (
    AuthenticationBackend,
    BearerTransport,
    JWTStrategy,
)
from fastapi_users.db import SQLAlchemyUserDatabase
from sqlalchemy.ext.asyncio import AsyncSession

from database import User, get_async_session

# Configure logger
logger = logging.getLogger(__name__)

# JWT Secret Key - Should be set via environment variable in production
SECRET = os.getenv("JWT_SECRET_KEY", "your-secret-key-change-in-production-please-use-a-strong-random-secret")
REFRESH_SECRET = os.getenv("JWT_REFRESH_SECRET_KEY", "your-refresh-secret-key-change-in-production-please-use-different-secret")


class UserManager(IntegerIDMixin, BaseUserManager[User, int]):
    """User manager for handling user operations"""
    
    reset_password_token_secret = SECRET
    verification_token_secret = SECRET

    async def on_after_register(self, user: User, request: Optional[Request] = None):
        """Hook called after user registration"""
        logger.info(f"User {user.id} has registered")
        # TODO: Send email with verification token

    async def on_after_forgot_password(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Hook called after forgot password request"""
        logger.info(f"Password reset requested for user {user.id}")
        # TODO: Send email with reset token (do not log the token)

    async def on_after_request_verify(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Hook called after verification request"""
        logger.info(f"Verification requested for user {user.id}")
        # TODO: Send email with verification token (do not log the token)


async def get_user_db(session: AsyncSession = Depends(get_async_session)):
    """Get user database adapter"""
    yield SQLAlchemyUserDatabase(session, User)


async def get_user_manager(user_db: SQLAlchemyUserDatabase = Depends(get_user_db)):
    """Get user manager instance"""
    yield UserManager(user_db)


def get_jwt_strategy() -> JWTStrategy:
    """Get JWT authentication strategy for access tokens"""
    return JWTStrategy(secret=SECRET, lifetime_seconds=3600)  # 1 hour access token


def get_refresh_jwt_strategy() -> JWTStrategy:
    """Get JWT authentication strategy for refresh tokens"""
    return JWTStrategy(secret=REFRESH_SECRET, lifetime_seconds=604800)  # 7 days refresh token


# Bearer transport for token-based authentication
bearer_transport = BearerTransport(tokenUrl="auth/jwt/login")

# Authentication backend with JWT for access tokens
auth_backend = AuthenticationBackend(
    name="jwt",
    transport=bearer_transport,
    get_strategy=get_jwt_strategy,
)

# Refresh token backend
refresh_bearer_transport = BearerTransport(tokenUrl="auth/jwt/refresh")

refresh_auth_backend = AuthenticationBackend(
    name="jwt-refresh",
    transport=refresh_bearer_transport,
    get_strategy=get_refresh_jwt_strategy,
)

# FastAPI Users instance
fastapi_users = FastAPIUsers[User, int](
    get_user_manager,
    [auth_backend, refresh_auth_backend],
)

# Current user dependency - for protected routes
current_active_user = fastapi_users.current_user(active=True)
current_verified_user = fastapi_users.current_user(active=True, verified=True)
