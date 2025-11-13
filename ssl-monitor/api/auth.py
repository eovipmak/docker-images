"""
Authentication configuration using FastAPI Users
"""
import os
from typing import Optional
from fastapi import Depends, Request
from fastapi_users import BaseUserManager, FastAPIUsers, IntegerIDMixin
from fastapi_users.authentication import (
    AuthenticationBackend,
    BearerTransport,
    JWTStrategy,
)
from fastapi_users.db import SQLAlchemyUserDatabase
from sqlalchemy.orm import Session

from database import User, get_db

# JWT Secret Key - Should be set via environment variable in production
SECRET = os.getenv("JWT_SECRET_KEY", "your-secret-key-change-in-production-please-use-a-strong-random-secret")


class UserManager(IntegerIDMixin, BaseUserManager[User, int]):
    """User manager for handling user operations"""
    
    reset_password_token_secret = SECRET
    verification_token_secret = SECRET

    async def on_after_register(self, user: User, request: Optional[Request] = None):
        """Hook called after user registration"""
        print(f"User {user.id} has registered with email {user.email}")

    async def on_after_forgot_password(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Hook called after forgot password request"""
        print(f"User {user.id} has forgotten their password. Reset token: {token}")
        # TODO: Send email with reset token

    async def on_after_request_verify(
        self, user: User, token: str, request: Optional[Request] = None
    ):
        """Hook called after verification request"""
        print(f"Verification requested for user {user.id}. Verification token: {token}")
        # TODO: Send email with verification token


async def get_user_db(db: Session = Depends(get_db)):
    """Get user database adapter"""
    yield SQLAlchemyUserDatabase(db, User)


async def get_user_manager(user_db: SQLAlchemyUserDatabase = Depends(get_user_db)):
    """Get user manager instance"""
    yield UserManager(user_db)


def get_jwt_strategy() -> JWTStrategy:
    """Get JWT authentication strategy with refresh token support"""
    return JWTStrategy(secret=SECRET, lifetime_seconds=3600)  # 1 hour access token


# Bearer transport for token-based authentication
bearer_transport = BearerTransport(tokenUrl="auth/jwt/login")

# Authentication backend with JWT
auth_backend = AuthenticationBackend(
    name="jwt",
    transport=bearer_transport,
    get_strategy=get_jwt_strategy,
)

# FastAPI Users instance
fastapi_users = FastAPIUsers[User, int](
    get_user_manager,
    [auth_backend],
)

# Current user dependency - for protected routes
current_active_user = fastapi_users.current_user(active=True)
current_verified_user = fastapi_users.current_user(active=True, verified=True)
