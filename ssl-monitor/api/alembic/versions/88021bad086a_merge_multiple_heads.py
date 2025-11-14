"""Merge multiple heads

Revision ID: 88021bad086a
Revises: 5f26bc034a6d, f9c12345abcd
Create Date: 2025-11-14 09:25:25.651406

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '88021bad086a'
down_revision: Union[str, None] = ('5f26bc034a6d', 'f9c12345abcd')
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    pass


def downgrade() -> None:
    pass
